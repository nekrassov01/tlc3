package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var (
	host = "localhost"
	port = "8443"
	addr = host + ":" + port
)

func getTime(value string, loc *time.Location) time.Time {
	t, _ := time.Parse(time.RFC3339, value)
	return t.In(loc)
}

func getNotBefore(loc *time.Location) time.Time {
	return getTime("2023-01-01T09:00:00+09:00", loc)
}

func getNotAfter(loc *time.Location) time.Time {
	return time.Now().Truncate(time.Hour).In(loc).Add(24 * time.Hour)
}

func getCurrentTime(loc *time.Location) time.Time {
	return time.Now().Truncate(time.Hour).In(loc)
}

func TestMain(m *testing.M) {
	server, tempDir, err := setup(addr)
	if err != nil {
		log.Fatal("failed to setup: ", err)
	}
	code := m.Run()
	if err := teardown(server, tempDir); err != nil {
		log.Fatal("failed to teardown: ", err)
	}
	os.Exit(code)
}

func setup(addr string) (*http.Server, string, error) {
	if err := setupEnv(); err != nil {
		return nil, "", fmt.Errorf("failed to set environment valiable: %w", err)
	}
	tempDir, certFile, keyFile, err := setupPath()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	if err := setupCert(certFile, keyFile); err != nil {
		return nil, "", fmt.Errorf("failed to create certificate: %w", err)
	}
	server := setupServer(addr)
	ch := make(chan error, 1)
	go func() {
		if err := server.ListenAndServeTLS(certFile, keyFile); err != http.ErrServerClosed {
			ch <- err
		}
		close(ch)
	}()
	if err := waitServer(addr, 5*time.Second); err != nil {
		return nil, "", fmt.Errorf("failed to start server: %w", err)
	}
	select {
	case err := <-ch:
		if err != nil {
			return nil, "", fmt.Errorf("failed to run server: %w", err)
		}
	default:
	}
	return server, tempDir, nil
}

func setupEnv() error {
	err := os.Setenv("TZ", "Asia/Tokyo")
	if err != nil {
		return err
	}
	return nil
}

func setupPath() (tempDir, certFile, keyFile string, err error) {
	tempDir, err = os.MkdirTemp("", "test")
	if err != nil {
		return "", "", "", err
	}
	return tempDir, filepath.Join(tempDir, "cert.pem"), filepath.Join(tempDir, "key.pem"), nil
}

func setupCert(certFile, keyFile string) error {
	// create private key
	privKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// configure certificate
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}
	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: "local test CA"},
		NotBefore:             getNotBefore(time.Local),
		NotAfter:              getNotAfter(time.Local),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &privKey.PublicKey, privKey)
	if err != nil {
		return err
	}

	// save certificate in pem
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	certOut.Close()

	// convert private key
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}

	// save private key in pem
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}
	keyOut.Close()

	return nil
}

func setupServer(addr string) *http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "test server")
	})
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, // #nosec G402
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server
}

func waitServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := tls.Dial("tcp", addr, &tls.Config{
			InsecureSkipVerify: true, // #nosec G402
		})
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start server in %v", timeout)
}

func teardown(server *http.Server, tempDir string) error {
	defer os.RemoveAll(tempDir)
	if err := server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("cannot shutdown server: %w", err)
	}
	return nil
}

func Test_getCertList(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx      context.Context
		addrs    []string
		timeout  time.Duration
		location *time.Location
		insecure bool
	}
	tests := []struct {
		name    string
		args    args
		want    []*certInfo
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				ctx:      ctx,
				addrs:    []string{addr},
				timeout:  5 * time.Second,
				location: time.Local,
				insecure: true,
			},
			want: []*certInfo{
				{
					DomainName:  host,
					AccessPort:  port,
					IPAddresses: []net.IP{net.ParseIP("::1"), net.ParseIP("127.0.0.1")},
					Issuer:      "CN=local test CA",
					CommonName:  "local test CA",
					SANs:        []string{},
					NotBefore:   getNotBefore(time.Local),
					NotAfter:    getNotAfter(time.Local),
					CurrentTime: getCurrentTime(time.Local),
					DaysLeft:    1,
				},
			},
			wantErr: false,
		},
		{
			name: "utc",
			args: args{
				ctx:      ctx,
				addrs:    []string{addr},
				timeout:  5 * time.Second,
				location: time.UTC,
				insecure: true,
			},
			want: []*certInfo{
				{
					DomainName:  host,
					AccessPort:  port,
					IPAddresses: []net.IP{net.ParseIP("::1"), net.ParseIP("127.0.0.1")},
					Issuer:      "CN=local test CA",
					CommonName:  "local test CA",
					SANs:        []string{},
					NotBefore:   getNotBefore(time.UTC),
					NotAfter:    getNotAfter(time.UTC),
					CurrentTime: getCurrentTime(time.UTC),
					DaysLeft:    1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCertList(tt.args.ctx, tt.args.addrs, tt.args.timeout, tt.args.insecure, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCertList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, g := range got {
				for _, w := range tt.want {
					if !reflect.DeepEqual(g.DomainName, w.DomainName) {
						t.Errorf("DoaminName = %v, want %v", g.DomainName, w.DomainName)
					}
					if !reflect.DeepEqual(g.AccessPort, w.AccessPort) {
						t.Errorf("AccessPort = %v, want %v", g.AccessPort, w.AccessPort)
					}
					if diff := cmp.Diff(g.IPAddresses, w.IPAddresses); diff != "" {
						t.Errorf(diff)
					}
					if !reflect.DeepEqual(g.Issuer, w.Issuer) {
						t.Errorf("Issuer = %v, want %v", g.Issuer, w.Issuer)
					}
					if !reflect.DeepEqual(g.CommonName, w.CommonName) {
						t.Errorf("CommonName = %v, want %v", g.CommonName, w.CommonName)
					}
					if !reflect.DeepEqual(g.SANs, w.SANs) {
						t.Errorf("SANs = %v, want %v", g.SANs, w.SANs)
					}
					if !reflect.DeepEqual(g.NotBefore, w.NotBefore) {
						t.Errorf("NotBefore = %v, want %v", g.NotBefore, w.NotBefore)
					}
					if !reflect.DeepEqual(g.NotAfter, w.NotAfter) {
						t.Errorf("NotAfter = %v, want %v", g.NotAfter, w.NotAfter)
					}
				}
			}
		})
	}
}

func Test_newConnector(t *testing.T) {
	type args struct {
		addr     string
		timeout  time.Duration
		location *time.Location
		insecure bool
	}
	tests := []struct {
		name    string
		args    args
		want    *connector
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				addr:     addr,
				timeout:  5 * time.Second,
				location: time.Local,
				insecure: false,
			},
			want: &connector{
				addr:     addr,
				host:     host,
				port:     port,
				timeout:  5 * time.Second,
				location: time.Local,
				tlsConfig: &tls.Config{
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: false, // #nosec G402
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newConnector(tt.args.addr, tt.args.timeout, tt.args.insecure, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("newConnector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newConnector() gotHost = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_connector_lookupIP(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		addr      string
		host      string
		port      string
		ips       []net.IP
		timeout   time.Duration
		tlsConfig *tls.Config
		tlsConn   *tls.Conn
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []net.IP
	}{
		{
			name: "basic",
			fields: fields{
				addr:      addr,
				host:      host,
				port:      port,
				ips:       nil,
				timeout:   5 * time.Second,
				tlsConfig: nil,
				tlsConn:   nil,
			},
			args: args{
				ctx: ctx,
			},
			want: []net.IP{net.ParseIP("::1"), net.ParseIP("127.0.0.1")},
		},
		{
			name: "empty",
			fields: fields{
				addr:      addr,
				host:      "dummy",
				port:      port,
				ips:       nil,
				timeout:   5 * time.Second,
				tlsConfig: nil,
				tlsConn:   nil,
			},
			args: args{
				ctx: ctx,
			},
			want: []net.IP{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipMap.Delete(tt.fields.host)
			c := &connector{
				addr:      tt.fields.addr,
				host:      tt.fields.host,
				port:      tt.fields.port,
				ips:       tt.fields.ips,
				timeout:   tt.fields.timeout,
				tlsConfig: tt.fields.tlsConfig,
				tlsConn:   tt.fields.tlsConn,
			}
			c.lookupIP(tt.args.ctx)
			if diff := cmp.Diff(c.ips, tt.want); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func Test_connector_getTLSConn(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		addr      string
		host      string
		port      string
		ips       []net.IP
		timeout   time.Duration
		tlsConfig *tls.Config
		tlsConn   *tls.Conn
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				addr:    addr,
				host:    host,
				port:    port,
				ips:     nil,
				timeout: 5 * time.Second,
				tlsConfig: &tls.Config{
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true, // #nosec G402
				},
				tlsConn: nil,
			},
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name: "error",
			fields: fields{
				addr:    ":443",
				host:    host,
				port:    port,
				ips:     nil,
				timeout: 5 * time.Second,
				tlsConfig: &tls.Config{
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true, // #nosec G402
				},
				tlsConn: nil,
			},
			args: args{
				ctx: ctx,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connMap.Delete(tt.fields.host)
			c := &connector{
				addr:      tt.fields.addr,
				host:      tt.fields.host,
				port:      tt.fields.port,
				ips:       tt.fields.ips,
				timeout:   tt.fields.timeout,
				tlsConfig: tt.fields.tlsConfig,
				tlsConn:   tt.fields.tlsConn,
			}
			if err := c.getTLSConn(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("connector.getTLSConn() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_connector_getServerCert(t *testing.T) {
	ctx := context.Background()
	type fields struct {
		addr      string
		host      string
		port      string
		ips       []net.IP
		timeout   time.Duration
		location  *time.Location
		tlsConfig *tls.Config
		tlsConn   *tls.Conn
	}
	tests := []struct {
		name    string
		fields  fields
		want    *certInfo
		wantErr bool
	}{
		{
			name: "basic",
			fields: fields{
				addr:     addr,
				host:     host,
				port:     port,
				ips:      []net.IP{},
				timeout:  5 * time.Second,
				location: time.Local,
				tlsConfig: &tls.Config{
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true, // #nosec G402
				},
			},
			want: &certInfo{
				DomainName:  host,
				AccessPort:  port,
				IPAddresses: []net.IP{},
				Issuer:      "CN=local test CA",
				CommonName:  "local test CA",
				SANs:        []string{},
				NotBefore:   getNotBefore(time.Local),
				NotAfter:    getNotAfter(time.Local),
				CurrentTime: getCurrentTime(time.Local),
				DaysLeft:    1,
			},
			wantErr: false,
		},
		{
			name: "utc",
			fields: fields{
				addr:     addr,
				host:     host,
				port:     port,
				ips:      []net.IP{},
				timeout:  5 * time.Second,
				location: time.UTC,
				tlsConfig: &tls.Config{
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true, // #nosec G402
				},
			},
			want: &certInfo{
				DomainName:  host,
				AccessPort:  port,
				IPAddresses: []net.IP{},
				Issuer:      "CN=local test CA",
				CommonName:  "local test CA",
				SANs:        []string{},
				NotBefore:   getNotBefore(time.UTC),
				NotAfter:    getNotAfter(time.UTC),
				CurrentTime: getCurrentTime(time.UTC),
				DaysLeft:    1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &connector{
				addr:      tt.fields.addr,
				host:      tt.fields.host,
				port:      tt.fields.port,
				ips:       tt.fields.ips,
				timeout:   tt.fields.timeout,
				location:  tt.fields.location,
				tlsConfig: tt.fields.tlsConfig,
				tlsConn:   tt.fields.tlsConn,
			}
			if err := c.getTLSConn(ctx); err != nil {
				t.Fatal(err)
			}
			got, err := c.getServerCert()
			if (err != nil) != tt.wantErr {
				t.Errorf("connector.getServerCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.DomainName, tt.want.DomainName) {
				t.Errorf("DoaminName = %v, want %v", got.DomainName, tt.want.DomainName)
			}
			if !reflect.DeepEqual(got.AccessPort, tt.want.AccessPort) {
				t.Errorf("AccessPort = %v, want %v", got.AccessPort, tt.want.AccessPort)
			}
			if !reflect.DeepEqual(got.IPAddresses, tt.want.IPAddresses) {
				t.Errorf("IPAddresses = %v, want %v", got.IPAddresses, tt.want.IPAddresses)
			}
			if !reflect.DeepEqual(got.Issuer, tt.want.Issuer) {
				t.Errorf("Issuer = %v, want %v", got.Issuer, tt.want.Issuer)
			}
			if !reflect.DeepEqual(got.CommonName, tt.want.CommonName) {
				t.Errorf("CommonName = %v, want %v", got.CommonName, tt.want.CommonName)
			}
			if !reflect.DeepEqual(got.SANs, tt.want.SANs) {
				t.Errorf("SANs = %v, want %v", got.SANs, tt.want.SANs)
			}
			if !reflect.DeepEqual(got.NotBefore, tt.want.NotBefore) {
				t.Errorf("NotBefore = %v, want %v", got.NotBefore, tt.want.NotBefore)
			}
			if !reflect.DeepEqual(got.NotAfter, tt.want.NotAfter) {
				t.Errorf("NotAfter = %v, want %v", got.NotAfter, tt.want.NotAfter)
			}
		})
	}
}

func Test_daysLeft(t *testing.T) {
	type args struct {
		notAfter time.Time
		now      time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "full year difference",
			args: args{
				notAfter: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 365,
		},
		{
			name: "one day before new year",
			args: args{
				notAfter: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "one day difference",
			args: args{
				notAfter: time.Date(2023, 1, 2, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "end of year",
			args: args{
				notAfter: time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: 364,
		},
		{
			name: "same day and time",
			args: args{
				notAfter: time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "one second difference",
			args: args{
				notAfter: time.Date(2023, 1, 1, 9, 0, 1, 0, time.UTC),
				now:      time.Date(2023, 1, 1, 9, 0, 0, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "less than one day",
			args: args{
				notAfter: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 12, 31, 9, 0, 1, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "last second",
			args: args{
				notAfter: time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC),
				now:      time.Date(2023, 12, 31, 8, 59, 59, 0, time.UTC),
			},
			want: 1,
		},
		{
			name: "leap year difference",
			args: args{
				notAfter: time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				now:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: 365,
		},
		{
			name: "end of leap year",
			args: args{
				notAfter: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
				now:      time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := daysLeft(tt.args.notAfter, tt.args.now); got != tt.want {
				t.Errorf("daysLeft() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ensureDefaultPort(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "basic",
			args: args{
				addr: addr,
			},
			want: addr,
		},
		{
			name: "default port",
			args: args{
				addr: host,
			},
			want: host + ":443",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ensureDefaultPort(tt.args.addr); got != tt.want {
				t.Errorf("ensureDefaultPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ensureHostPort(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name     string
		args     args
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{
			name: "basic",
			args: args{
				addr: addr,
			},
			wantHost: host,
			wantPort: port,
			wantErr:  false,
		},
		{
			name: "invalid addr error",
			args: args{
				addr: "localhost::443",
			},
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
		{
			name: "port out of range error",
			args: args{
				addr: "localhost:65536",
			},
			wantHost: "",
			wantPort: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, err := ensureHostPort(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ensureHostPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHost != tt.wantHost {
				t.Errorf("ensureHostPort() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("ensureHostPort() gotPort = %v, want %v", gotPort, tt.wantPort)
			}
		})
	}
}

func Test_getSANs(t *testing.T) {
	uri, err := url.Parse("example.com/")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		cert *x509.Certificate
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "basic",
			args: args{
				cert: &x509.Certificate{
					DNSNames:       []string{"example.com", "www.example.com"},
					EmailAddresses: []string{"aaa@example.com", "bbb@example.com"},
					IPAddresses:    []net.IP{net.ParseIP("192.168.10.10"), net.ParseIP("192.168.10.20")},
					URIs:           []*url.URL{uri},
				},
			},
			want: []string{"example.com", "www.example.com", "aaa@example.com", "bbb@example.com", "192.168.10.10", "192.168.10.20", "example.com/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSANs(tt.args.cert); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSANs() = %v, want %v", got, tt.want)
			}
		})
	}
}
