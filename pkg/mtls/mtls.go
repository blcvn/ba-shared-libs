package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

type CertReloader struct {
	mu       sync.RWMutex
	cert     *tls.Certificate
	certPath string
	keyPath  string
	pool     *x509.CertPool
}

func NewCertReloader(certPath, keyPath string) (*CertReloader, error) {
	reloader := &CertReloader{
		certPath: certPath,
		keyPath:  keyPath,
	}
	// Initial load with retry to handle Vault Agent race condition
	var err error
	for i := 0; i < 10; i++ {
		if err = reloader.reload(); err == nil {
			go reloader.watch()
			return reloader, nil
		}
		fmt.Printf("MtLS: Waiting for certificates %s... (%d/10)\n", certPath, i+1)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("initial load failed after retries: %v", err)
}

func (r *CertReloader) reload() error {
	// Load KeyPair
	cert, err := tls.LoadX509KeyPair(r.certPath, r.keyPath)
	if err != nil {
		return fmt.Errorf("LoadX509KeyPair: %v", err)
	}

	// Load CA for Client Auth
	pemData, err := ioutil.ReadFile(r.certPath)
	if err != nil {
		return fmt.Errorf("ReadFile: %v", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemData) {
		return fmt.Errorf("failed to append certs from %s", r.certPath)
	}

	r.mu.Lock()
	r.cert = &cert
	r.pool = pool
	r.mu.Unlock()
	return nil
}

func (r *CertReloader) watch() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30s
	for range ticker.C {
		if err := r.reload(); err != nil {
			log.Printf("MtLS: Failed to reload certs: %v", err)
		} else {
			// Log only on success if verbose? Or just silent.
		}
	}
}

// GetConfigForClient implements tls.Config.GetConfigForClient to return dynamic config for verify incoming client certs (Server Side)
func (r *CertReloader) GetConfigForClient(hello *tls.ClientHelloInfo) (*tls.Config, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return &tls.Config{
		Certificates: []tls.Certificate{*r.cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    r.pool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// GetClientCertificate implements tls.Config.GetClientCertificate for outgoing mTLS (Client Side)
func (r *CertReloader) GetClientCertificate(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cert, nil
}

// ClientTLSConfig returns a tls.Config for use in http.Transport (Client Side)
func (r *CertReloader) ClientTLSConfig() *tls.Config {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// We need RootCAs to verify the Server we are connecting to (e.g. Kratos)
	// Assuming the Server uses the same CA as we do (Internal PKI)
	return &tls.Config{
		GetClientCertificate: r.GetClientCertificate,
		RootCAs:              r.pool,
		MinVersion:           tls.VersionTLS12,
	}
}
