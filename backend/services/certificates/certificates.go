package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

const certsFile = "/certs.json"

// Config for the CertificateService.
type Config struct {
	CertificateStoreDirectory string
	AcmeEndpoint              string
	AcmeEmail                 string
	AcmeHttpPort              string
	AcmeTlsPort               string
}

// CertificateService stores and manages certificates.
type CertificateService struct {
	config       *Config
	certificates map[string]*CertificateResource
	privateKey   *rsa.PrivateKey
	acmeHelper   *ACMEHelper
}

// NewCertificateService creates a new CertificateService with the given config parameters and load the existing certificates from disk.
func NewCertificateService(certificateStoreDirectory string, acmeEndpoint string, acmeEmail string) (*CertificateService, error) {
	certificateStore := &CertificateService{
		config: &Config{
			CertificateStoreDirectory: certificateStoreDirectory,
			AcmeEndpoint:              acmeEndpoint,
			AcmeEmail:                 acmeEmail,
			AcmeHttpPort:              "80",
			AcmeTlsPort:               "443",
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
	_, err = certificateStore.GetOrGeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("couldn't get or generate private key: %w", err)
	}

	// Initialize the ACME helper
	acmeHelper, err := NewACMEHelper(certificateStore.privateKey, acmeEmail, acmeEndpoint, certificateStore.config.AcmeHttpPort, certificateStore.config.AcmeTlsPort)
	if err != nil {
		return nil, err
	}
	certificateStore.acmeHelper = acmeHelper

	return certificateStore, nil
}

// Get returns a certificate from the store.
func (s *CertificateService) Get(domain string) (*CertificateResource, error) {
	cert, ok := s.certificates[domain]
	if !ok {
		return nil, fmt.Errorf("couldn't find certificate for domain %s", domain)
	}
	return cert, nil
}

// Add saves a certificate to the store.
func (s *CertificateService) Add(domain string, cert *CertificateResource) error {
	s.certificates[domain] = cert

	err := s.saveCertificatesToFile()
	if err != nil {
		return fmt.Errorf("couldn't save certificate to disk: %w", err)
	}

	return nil
}

// GetOrCreateCertificate gets a certificate from the store and creates a new one if needed (and saves it).
func (s *CertificateService) GetOrCreateCertificate(domain string) (*CertificateResource, error) {

	cert, err := s.Get(domain)
	if err == nil && cert != nil {
		return cert, nil
	}

	cert, err = s.acmeHelper.generateCertificateWithACMEChallenge(domain)
	if err != nil {
		return nil, fmt.Errorf("couldn't create certificate: %w", err)
	}

	err = s.Add(domain, cert)
	if err != nil {
		return cert, fmt.Errorf("couldn't save certificate: %w", err)
	}

	return cert, nil

}

// GetOrGeneratePrivateKey gets the private key from the store or generates and saves a new one.
func (s *CertificateService) GetOrGeneratePrivateKey() (*rsa.PrivateKey, error) {
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

	for domain, cert := range s.certificates {
		certFileName := fmt.Sprintf("%s/%s.cert.pem", s.config.CertificateStoreDirectory, domain)
		err := os.WriteFile(certFileName, cert.Certificate, 0644)
		if err != nil {
			return fmt.Errorf("couldn't save certificate to disk for domain %s: %w", domain, err)
		}
		cert.CertificateFile = certFileName

		keyFileName := fmt.Sprintf("%s/%s.private.key", s.config.CertificateStoreDirectory, domain)
		err = os.WriteFile(keyFileName, cert.PrivateKey, 0644)
		if err != nil {
			return fmt.Errorf("couldn't save private key to disk for domain %s: %w", domain, err)
		}
		cert.PrivateKeyFile = keyFileName
	}

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
