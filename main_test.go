package tlc3

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	loc, _         = time.LoadLocation("Asia/Tokyo")
	host           = "localhost"
	port           = "8443"
	addr           = host + ":" + port
	ipAddresses    = []netip.Addr{netip.MustParseAddr("127.0.0.1"), netip.MustParseAddr("::1")}
	issuer         = "CN=local test CA"
	commonName     = "local test CA"
	sans           = []string{"localhost", "127.0.0.1"}
	notBefore      = mustTime("2024-01-01T09:00:00+09:00", loc)
	notAfter       = mustTime("2025-01-02T09:00:00+09:00", loc)
	currentTime    = mustTime("2025-01-01T09:00:00+09:00", loc)
	notBeforeUTC   = mustTime("2024-01-01T00:00:00Z", time.UTC)
	notAfterUTC    = mustTime("2025-01-02T00:00:00Z", time.UTC)
	currentTimeUTC = mustTime("2025-01-01T00:00:00Z", time.UTC)
	daysLeft       = 1
)

var (
	basicExpects = []*CertInfo{
		{
			DomainName:  host,
			AccessPort:  port,
			IPAddresses: ipAddresses,
			Issuer:      issuer,
			CommonName:  commonName,
			SANs:        []string{},
			NotBefore:   notBefore,
			NotAfter:    notAfter,
			CurrentTime: currentTime,
			DaysLeft:    daysLeft,
		},
	}
	utcExpects = []*CertInfo{
		{
			DomainName:  host,
			AccessPort:  port,
			IPAddresses: ipAddresses,
			Issuer:      issuer,
			CommonName:  commonName,
			SANs:        []string{},
			NotBefore:   notBeforeUTC,
			NotAfter:    notAfterUTC,
			CurrentTime: currentTimeUTC,
			DaysLeft:    1,
		},
	}
	renderInput = []*CertInfo{
		{
			DomainName:  host,
			AccessPort:  port,
			IPAddresses: ipAddresses,
			Issuer:      issuer,
			CommonName:  commonName,
			SANs:        sans,
			NotBefore:   notBefore,
			NotAfter:    notAfter,
			CurrentTime: currentTime,
			DaysLeft:    daysLeft,
		},
	}
)

func TestMain(m *testing.M) {
	nowFunc = func() time.Time {
		return currentTime
	}
	server, tempDir, err := setup(addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := teardown(server, tempDir); err != nil {
			panic(err)
		}
		nowFunc = time.Now
	}()
	m.Run()
}

func mustTime(s string, loc *time.Location) time.Time {
	t, err := time.ParseInLocation(time.RFC3339, s, loc)
	if err != nil {
		panic(err)
	}
	return t
}

func setup(addr string) (*http.Server, string, error) {
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
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             notBeforeUTC,
		NotAfter:              notAfterUTC,
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
	certOut, err := os.Create(filepath.Clean(certFile))
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err := certOut.Close(); err != nil {
		return err
	}

	// convert private key
	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return err
	}

	// save private key in pem
	keyOut, err := os.Create(filepath.Clean(keyFile))
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}
	if err := keyOut.Close(); err != nil {
		return err
	}

	return nil
}

func setupServer(addr string) *http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
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
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start server in %v", timeout)
}

func teardown(server *http.Server, tempDir string) error {
	err1 := server.Shutdown(context.Background())
	err2 := os.RemoveAll(tempDir)
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	return errors.Join(err1, err2)
}
