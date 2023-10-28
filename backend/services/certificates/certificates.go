package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

const certsFile = "/certs.json"

var certificatesLock = sync.RWMutex{}

// Config for the CertificateService.
type Config struct {
	CertificateStoreDirectory     string
	AcmeEndpoint                  string
	AcmeEmail                     string
	AcmeHttpPort                  string
	AcmeTlsPort                   string
	CertificateRenewValidDuration time.Duration
	CertificateRenewInterval      time.Duration
}

// CertificateService stores and manages certificates.
type CertificateService struct {
	config       *Config
	certificates map[string]*CertificateResource
	privateKey   *rsa.PrivateKey
	acmeHelper   ACME
}

const (
	DefaultAcmeHttpPort                  = "80"
	DefaultAcmeTlsPort                   = "443"
	DefaultCertificateRenewValidDuration = time.Hour * 24 * 30 // 30 days = Let's Encrypt default
	DefaultCertificateRenewInterval      = time.Hour * 24
)

// NewCertificateService creates a new CertificateService with the given config parameters and load the existing certificates from disk.
func NewCertificateService(certificateStoreDirectory string, acmeEndpoint string, acmeEmail string) (*CertificateService, error) {
	certificateStore := &CertificateService{
		config: &Config{
			CertificateStoreDirectory:     certificateStoreDirectory,
			AcmeEndpoint:                  acmeEndpoint,
			AcmeEmail:                     acmeEmail,
			AcmeHttpPort:                  DefaultAcmeHttpPort,
			AcmeTlsPort:                   DefaultAcmeTlsPort,
			CertificateRenewValidDuration: DefaultCertificateRenewValidDuration,
			CertificateRenewInterval:      DefaultCertificateRenewInterval,
		},
		certificates: map[string]*CertificateResource{},
	}

	if _, err := os.Stat(certificateStoreDirectory + certsFile); errors.Is(err, os.ErrNotExist) {
		err := os.WriteFile(certificateStoreDirectory+certsFile, []byte("{}"), 0644)
		if err != nil {
			return nil, fmt.Errorf("couldn't initizalize empty certificate service on disk: %w", err)
		}
	}

	err := certificateStore.loadCertificatesFromFile()
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize certificate service with certificates from disk: %w", err)
	}

	// Load private key
	_, err = certificateStore.getOrGeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("couldn't get or generate private key: %w", err)
	}

	// Initialize the ACME helper
	acmeHelper, err := NewACMEHelper(certificateStore.privateKey, acmeEmail, acmeEndpoint, certificateStore.config.AcmeHttpPort, certificateStore.config.AcmeTlsPort)
	if err != nil {
		return nil, err
	}
	certificateStore.acmeHelper = acmeHelper

	certificateStore.startRenewCertificateProcess()

	return certificateStore, nil
}

// Get returns a certificate from the store.
func (s *CertificateService) Get(domain string) (*CertificateResource, error) {
	certificatesLock.RLock()
	defer certificatesLock.RUnlock()
	cert, ok := s.certificates[domain]
	if !ok {
		return nil, fmt.Errorf("couldn't find certificate for domain %s", domain)
	}
	return cert, nil
}

// Add saves a certificate to the store.
func (s *CertificateService) Add(domain string, cert *CertificateResource) error {
	certificatesLock.Lock()
	defer certificatesLock.Unlock()

	s.certificates[domain] = cert

	err := s.saveCertificatesToFile()
	if err != nil {
		return fmt.Errorf("couldn't save certificate to disk: %w", err)
	}

	return nil
}

// certificateResourceToCertificate converts a CertificateResource to a tls.Certificate
func certificateResourceToCertificate(certResource *CertificateResource) (*tls.Certificate, error) {

	cert, err := tls.X509KeyPair(certResource.Certificate, certResource.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &cert, nil

}

// GetOrCreateTlsConfig creates a tls.Config for a domain.
// the tls.Config will always get (or create) the certificate from the the certificate service.
func (c *CertificateService) GetOrCreateTlsConfig(domain string) (*tls.Config, error) {

	// Get the certificate resource at creation time, so that we can fail fast on the initial setup.
	certResource, err := c.getOrCreateCertificateResource(domain)
	if err != nil {
		return nil, err
	}

	_, err = certificateResourceToCertificate(certResource)
	if err != nil {
		return nil, err
	}

	tlsConfig := tls.Config{
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			certRsource, err := c.getOrCreateCertificateResource(domain)
			if err != nil {
				return nil, err
			}
			return certificateResourceToCertificate(certRsource)
		},
	}

	return &tlsConfig, nil
}

// getOrCreateCertificateResource gets a certificate from the store and creates a new one if needed (and saves it).
func (s *CertificateService) getOrCreateCertificateResource(domain string) (*CertificateResource, error) {

	cert, err := s.Get(domain)
	if err == nil && cert != nil {
		return cert, nil
	}

	cert, err = s.acmeHelper.GenerateCertificateWithACMEChallenge(domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't create certificate: %w", err)
	}

	err = s.Add(domain, cert)
	if err != nil {
		return cert, fmt.Errorf("couldn't save certificate: %w", err)
	}

	return cert, nil

}

// startRenewCertificateProcess checks all certificates each CertificateRenewInterval
// and renews those that expire in less than CertificateRenewValidDuration
func (s *CertificateService) startRenewCertificateProcess() {

	go func() {

		ticker := time.NewTicker(s.config.CertificateRenewInterval)

		for range ticker.C {

			for _, domain := range s.getAllDomains() {

				cert, err := s.Get(domain)
				if err != nil {
					log.Errorf("couldn't get certificate for %s: %v", domain, err)
				}

				// renew certificate when needed.
				if s.config.CertificateRenewValidDuration < time.Until(cert.NotValidAfter) {

					err := s.renewCertificate(domain)
					if err != nil {
						log.Errorf("couldn't renew certificate for %s: %v", domain, err)
					} else {
						log.Printf("renewed certificate for domain: %s", domain)
					}
				}
			}

		}

	}()

}

// getAllDomains returns a slice of all domains in the certificate store.
func (s *CertificateService) getAllDomains() []string {
	certificatesLock.RLock()
	defer certificatesLock.RUnlock()
	return maps.Keys(s.certificates)
}

// renewCertificate renews a certificate.
func (s *CertificateService) renewCertificate(domain string) error {

	cert, err := s.acmeHelper.GenerateCertificateWithACMEChallenge(domain)
	if err != nil {
		return fmt.Errorf("couldn't create certificate: %w", err)
	}

	err = s.Add(domain, cert)
	if err != nil {
		return fmt.Errorf("couldn't save certificate: %w", err)
	}

	return nil
}

// getOrGeneratePrivateKey gets the private key from the store or generates and saves a new one.
func (s *CertificateService) getOrGeneratePrivateKey() (*rsa.PrivateKey, error) {
	if s.privateKey != nil {
		return s.privateKey, nil
	}

	key, err := s.generatePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("couldn't generate private key: %w", err)
	}
	s.privateKey = key

	return key, nil
}

// generatePrivateKey generates a private key.
func (s *CertificateService) generatePrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// certificatesFile is a helper struct to save the private key and the certificates to disk.
type certificatesFile struct {
	Certificates map[string]*CertificateResource
	PrivateKey   *rsa.PrivateKey
}

// saveCertificatesToFile saves the certificates and the private key to disk.
func (s *CertificateService) saveCertificatesToFile() error {

	certificatesFile := certificatesFile{
		Certificates: s.certificates,
		PrivateKey:   s.privateKey,
	}

	jsonData, err := json.MarshalIndent(certificatesFile, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	err = os.WriteFile(s.config.CertificateStoreDirectory+certsFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

// loadCertificatesFromFile loads the certificates and the private key from disk.
func (s *CertificateService) loadCertificatesFromFile() error {

	jsonData, err := os.ReadFile(s.config.CertificateStoreDirectory + certsFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	certificatesFile := &certificatesFile{
		Certificates: map[string]*CertificateResource{},
	}

	err = json.Unmarshal(jsonData, certificatesFile)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	s.certificates = certificatesFile.Certificates
	s.privateKey = certificatesFile.PrivateKey

	return nil
}
