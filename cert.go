package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"runtime"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type certInfo struct {
	DomainName  string
	AccessPort  string
	IPAddresses []string
	Issuer      string
	CommonName  string
	SANs        []string
	NotBefore   time.Time
	NotAfter    time.Time
	CurrentTime time.Time
	DaysLeft    int
}

func getCertList(ctx context.Context, addrs []string, timeout time.Duration, insecure bool, location *time.Location) ([]*certInfo, error) {
	res := make([]*certInfo, len(addrs))
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	eg, ctx := errgroup.WithContext(ctx)
	for i, addr := range addrs {
		i, addr := i, addr
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		eg.Go(func() error {
			defer sem.Release(1)
			conn, err := newConnector(addr, timeout, insecure, location)
			if err != nil {
				return err
			}
			if err := conn.getTLSConn(ctx); err != nil {
				return err
			}
			conn.lookupIP(ctx)
			info, err := conn.getServerCert()
			if err != nil {
				return err
			}
			defer conn.tlsConn.Close()
			res[i] = info
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

type connector struct {
	addr      string
	host      string
	port      string
	ips       []string
	timeout   time.Duration
	location  *time.Location
	tlsConfig *tls.Config
	tlsConn   *tls.Conn
}

func newConnector(addr string, timeout time.Duration, insecure bool, location *time.Location) (*connector, error) {
	addr = ensureDefaultPort(addr)
	host, port, err := ensureHostPort(addr)
	if err != nil {
		return nil, err
	}
	conn := &connector{
		addr:     addr,
		host:     host,
		port:     port,
		timeout:  timeout,
		location: location,
		tlsConfig: &tls.Config{
			ServerName:         host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure, // #nosec G402
		},
	}
	return conn, nil
}

// Since IP address lookup is not the primary responsibility of this application,
// it does not return an error but only a zero value in case of failure.
func (c *connector) lookupIP(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resolver := net.Resolver{}
	ips, err := resolver.LookupIP(ctx, "ip", c.host)
	if err != nil {
		c.ips = []string{}
	}
	for _, ip := range ips {
		c.ips = append(c.ips, ip.String())
	}
	slices.Sort(c.ips)
}

func (c *connector) getTLSConn(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	dialer := tls.Dialer{Config: c.tlsConfig}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("cannot connect to \"%s\": %w", c.addr, err)
	}
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return fmt.Errorf("connection is not TLS")
	}
	c.tlsConn = tlsConn
	return nil
}

func (c *connector) getServerCert() (*certInfo, error) {
	certs := c.tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("cannot find cert for \"%s\"", c.host)
	}
	cert := certs[0]
	now := time.Now()
	info := &certInfo{
		DomainName:  c.host,
		AccessPort:  c.port,
		IPAddresses: c.ips,
		Issuer:      cert.Issuer.String(),
		CommonName:  cert.Subject.CommonName,
		SANs:        getSANs(cert),
		NotBefore:   cert.NotBefore.In(c.location),
		NotAfter:    cert.NotAfter.In(c.location),
		CurrentTime: now.In(c.location).Truncate(time.Second),
		DaysLeft:    daysLeft(cert.NotAfter, now),
	}
	return info, nil
}

func daysLeft(t time.Time, u time.Time) int {
	return int(t.Sub(u).Hours() / 24)
}

func ensureDefaultPort(addr string) string {
	if !strings.Contains(addr, ":") {
		addr += ":443"
	}
	return addr
}

func ensureHostPort(addr string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return "", "", err
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return "", "", err
	}
	return host, port, nil
}

func getSANs(cert *x509.Certificate) []string {
	sans := make([]string, 0, len(cert.DNSNames)+len(cert.EmailAddresses)+len(cert.IPAddresses)+len(cert.URIs))
	sans = append(sans, cert.DNSNames...)
	sans = append(sans, cert.EmailAddresses...)
	for _, ip := range cert.IPAddresses {
		sans = append(sans, ip.String())
	}
	for _, uri := range cert.URIs {
		sans = append(sans, uri.String())
	}
	return sans
}
