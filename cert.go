package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"runtime"
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
	NotBefore   string
	NotAfter    string
	CurrentTime string
	DaysLeft    int
}

func getCertList(ctx context.Context, addrs []string, timeout string, insecure bool) ([]*certInfo, error) {
	duration, err := time.ParseDuration(timeout)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot convert \"%s\" to time duration: valid time units are ns|us|ms|s|m|h: %w",
			timeout,
			err,
		)
	}
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
			addr = ensureDefaultPort(addr)
			host, port, err := ensureHostPort(addr)
			if err != nil {
				return err
			}
			params := newParams(addr, host, port, duration, insecure)
			params.lookupIP(ctx)
			if err := params.getTLSConn(ctx); err != nil {
				return err
			}
			info, err := params.getServerCert()
			if err != nil {
				return err
			}
			defer params.tlsConn.Close()
			res[i] = info
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

type params struct {
	addr      string
	host      string
	port      string
	ips       []string
	timeout   time.Duration
	tlsConfig *tls.Config
	tlsConn   *tls.Conn
}

func newParams(addr, host, port string, timeout time.Duration, insecure bool) *params {
	return &params{
		addr:    addr,
		host:    host,
		port:    port,
		timeout: timeout,
		tlsConfig: &tls.Config{
			ServerName:         host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure, // #nosec G402
		},
	}
}

// Since IP address lookup is not the primary responsibility of this application,
// it does not return an error but only a zero value in case of failure.
func (p *params) lookupIP(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	resolver := net.Resolver{}
	ips, err := resolver.LookupIP(ctx, "ip", p.host)
	if err != nil {
		p.ips = []string{}
	}
	for _, ip := range ips {
		p.ips = append(p.ips, ip.String())
	}
}

func (p *params) getTLSConn(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	dialer := tls.Dialer{Config: p.tlsConfig}
	conn, err := dialer.DialContext(ctx, "tcp", p.addr)
	if err != nil {
		return fmt.Errorf("cannot connect to \"%s\": %w", p.addr, err)
	}
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return fmt.Errorf("connection is not TLS")
	}
	p.tlsConn = tlsConn
	return nil
}

func (p *params) getServerCert() (*certInfo, error) {
	certs := p.tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("cannot find cert for \"%s\"", p.host)
	}
	cert := certs[0]
	now := time.Now().Truncate(time.Minute)
	return &certInfo{
		DomainName:  p.host,
		AccessPort:  p.port,
		IPAddresses: p.ips,
		Issuer:      cert.Issuer.String(),
		CommonName:  cert.Subject.CommonName,
		SANs:        getSANs(cert),
		NotBefore:   formatTime(cert.NotBefore),
		NotAfter:    formatTime(cert.NotAfter),
		CurrentTime: formatTime(now),
		DaysLeft:    int(cert.NotAfter.Sub(now).Hours() / 24),
	}, nil
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
		return "", "", fmt.Errorf("cannot split host:port for \"%s\": %w", addr, err)
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return "", "", fmt.Errorf("invalid port detected on \"%s\": %w", addr, err)
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

func formatTime(t time.Time) string {
	return t.In(time.Local).Format("2006-01-02T15:04:05-07:00")
}
