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

func getNotBefore() string {
	form := "2006-01-02T15:04:05-07:00"
	base := "2023-01-01T09:00:00+90:00"
	t, err := time.Parse(form, base)
	if err != nil {
		return base
	}
	loc := t.In(time.Local)
	return loc.Format(form)
}

func getNotAfter() string {
	return time.Now().Truncate(time.Minute).In(time.Local).Add(24 * time.Hour).Format("2006-01-02T15:04:05-07:00")
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
	notBefore, err := time.Parse(time.RFC3339, getNotBefore())
	if err != nil {
		return err
	}
	notAfter, err := time.Parse(time.RFC3339, getNotAfter())
	if err != nil {
		return err
	}
	tmpl := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{CommonName: "local test CA"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
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
				insecure: true,
			},
			want: []*certInfo{
				{
					DomainName:  host,
					AccessPort:  port,
					IPAddresses: []string{"127.0.0.1", "::1"},
					Issuer:      "CN=local test CA",
					CommonName:  "local test CA",
					SANs:        []string{},
					NotBefore:   getNotBefore(),
					NotAfter:    getNotAfter(),
					CurrentTime: formatTime(time.Now().Truncate(time.Minute)),
					DaysLeft:    1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getCertList(tt.args.ctx, tt.args.addrs, tt.args.timeout, tt.args.insecure)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCertList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCertList() = %v, want %v", got, tt.want)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func Test_newConnector(t *testing.T) {
	type args struct {
		addr     string
		host     string
		port     string
		timeout  time.Duration
		insecure bool
	}
	tests := []struct {
		name string
		args args
		want *connector
	}{
		{
			name: "basic",
			args: args{
				addr:     addr,
				host:     host,
				port:     port,
				timeout:  5 * time.Second,
				insecure: false,
			},
			want: &connector{
				addr:    addr,
				host:    host,
				port:    port,
				timeout: 5 * time.Second,
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
			if got := newConnector(tt.args.addr, tt.args.host, tt.args.port, tt.args.timeout, tt.args.insecure); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newConnector() = %v, want %v", got, tt.want)
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
		ips       []string
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
		want   []string
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
			want: []string{"127.0.0.1", "::1"},
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
			want: []string{},
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
				tlsConfig: tt.fields.tlsConfig,
				tlsConn:   tt.fields.tlsConn,
			}
			c.lookupIP(tt.args.ctx)
			if !reflect.DeepEqual(c.ips, tt.want) {
				t.Errorf("connector.lookupIP() = %v, want %v", c.ips, tt.want)
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
		ips       []string
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
		ips       []string
		timeout   time.Duration
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
				addr:    addr,
				host:    host,
				port:    port,
				ips:     []string{},
				timeout: 5 * time.Second,
				tlsConfig: &tls.Config{
					ServerName:         host,
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true, // #nosec G402
				},
			},
			want: &certInfo{
				DomainName:  host,
				AccessPort:  port,
				IPAddresses: []string{},
				Issuer:      "CN=local test CA",
				CommonName:  "local test CA",
				SANs:        []string{},
				NotBefore:   getNotBefore(),
				NotAfter:    getNotAfter(),
				CurrentTime: formatTime(time.Now().Truncate(time.Minute)),
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("connector.getServerCert() = %v, want %v", got, tt.want)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf(diff)
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
