package tlc3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var (
	nowFunc = time.Now
	ipMap   sync.Map
	connMap sync.Map
)

// CertInfo represents certificate information.
type CertInfo struct {
	DomainName  string
	AccessPort  string
	IPAddresses []netip.Addr
	Issuer      string
	CommonName  string
	SANs        []string
	NotBefore   time.Time
	NotAfter    time.Time
	CurrentTime time.Time
	DaysLeft    int
}

// GetCerts retrieves certificate information for the given addresses.
func GetCerts(ctx context.Context, addrs []string, timeout time.Duration, insecure bool, location *time.Location) ([]*CertInfo, error) {
	result := make([]*CertInfo, len(addrs))

	// function to process each address
	fn := func(i int, addr string) error {
		conn, err := newConnector(addr, timeout, insecure, location)
		if err != nil {
			return err
		}
		if err := conn.connect(ctx); err != nil {
			return err
		}
		defer conn.release()
		conn.lookupIP(ctx)
		info, err := conn.getCert()
		if err != nil {
			return err
		}
		result[i] = info
		return nil
	}

	// when only one address is provided, process it directly
	if len(addrs) == 1 {
		if err := fn(0, addrs[0]); err != nil {
			return nil, err
		}
		return result[:1], nil
	}

	// process multiple addresses concurrently
	sem := semaphore.NewWeighted(int64(runtime.NumCPU()))
	eg, ctx := errgroup.WithContext(ctx)
	for i, addr := range addrs {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		i := i
		addr := addr
		eg.Go(func() error {
			defer sem.Release(1)
			return fn(i, addr)
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

// connector handles TLS connection and certificate retrieval.
type connector struct {
	addr     string
	host     string
	port     string
	ips      []netip.Addr
	timeout  time.Duration
	location *time.Location
	config   *tls.Config
	conn     *tls.Conn
	mu       sync.Mutex
}

// newConnector creates a new connector for the given address.
func newConnector(addr string, timeout time.Duration, insecure bool, location *time.Location) (*connector, error) {
	if !strings.Contains(addr, ":") {
		addr += ":443"
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	conn := &connector{
		addr:     addr,
		host:     host,
		port:     port,
		timeout:  timeout,
		location: location,
		config: &tls.Config{
			ServerName:         host,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure, // #nosec G402
		},
	}
	return conn, nil
}

// connect establishes a TLS connection to the server.
func (c *connector) connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if conn, ok := connMap.Load(c.host); ok {
		c.conn = conn.(*tls.Conn)
		return nil
	}
	dialer := tls.Dialer{
		Config:    c.config,
		NetDialer: &net.Dialer{Timeout: c.timeout},
	}
	conn, err := dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("cannot connect to %q: %w", c.addr, err)
	}
	var ok bool
	c.conn, ok = conn.(*tls.Conn)
	if !ok {
		conn.Close()
		return fmt.Errorf("connection is not TLS")
	}
	connMap.Store(c.host, c.conn)
	return nil
}

// release releases the TLS connection back to the connection pool.
func (c *connector) release() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		connMap.Store(c.host, c.conn)
		c.conn = nil
	}
}

// getCert retrieves the certificate information from the TLS connection.
func (c *connector) getCert() (*CertInfo, error) {
	if c == nil || c.conn == nil {
		return nil, fmt.Errorf("cannot find connection for %q", c.host)
	}
	certs := c.conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return nil, fmt.Errorf("cannot find cert for %q", c.host)
	}
	cert := certs[0]
	now := nowFunc()
	info := &CertInfo{
		DomainName:  c.host,
		AccessPort:  c.port,
		IPAddresses: c.ips,
		Issuer:      cert.Issuer.String(),
		CommonName:  cert.Subject.CommonName,
		SANs:        getSANs(cert),
		NotBefore:   cert.NotBefore.In(c.location),
		NotAfter:    cert.NotAfter.In(c.location),
		CurrentTime: now.In(c.location).Truncate(time.Second),
		DaysLeft:    getDaysLeft(cert.NotAfter, now),
	}
	return info, nil
}

// lookupIP looks up IP addresses for the host.
// Since IP address lookup is not the primary responsibility of this application,
// it does not return an error but only a zero value in case of failure.
func (c *connector) lookupIP(ctx context.Context) {
	if caches, ok := ipMap.Load(c.host); ok {
		c.ips = caches.([]netip.Addr)
		return
	}
	var resolver net.Resolver
	var err error
	c.ips, err = resolver.LookupNetIP(ctx, "ip", c.host)
	if err != nil {
		c.ips = nil
	}
	slices.SortFunc(c.ips, func(a, b netip.Addr) int {
		return a.Compare(b)
	})
	ipMap.Store(c.host, c.ips)
}

// getSANs extracts Subject Alternative Names from the certificate.
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

// getDaysLeft calculates the number of days left until the given time.
func getDaysLeft(t time.Time, u time.Time) int {
	return int(t.Sub(u).Hours() / 24)
}
