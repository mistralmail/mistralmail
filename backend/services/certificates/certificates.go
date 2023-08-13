package certificates

import (
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
	// AcmeHttpPort         string
	// AcmeTlsPort          string
}

// CertificateService stores and manages certificates.
type CertificateService struct {
	config       *Config
	certificates map[string]*CertificateResource
}

// NewCertificateService creates a new CertificateService with the given config parameters and load the existing certificates from disk.
func NewCertificateService(certificateStoreDirectory string, acmeEndpoint string, acmeEmail string) (*CertificateService, error) {
	certificateStore := &CertificateService{
		config: &Config{
			CertificateStoreDirectory: certificateStoreDirectory,
			AcmeEndpoint:              acmeEndpoint,
			AcmeEmail:                 acmeEmail,
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

	cert, err = generateCertificateWithACMEChallenge(domain, s.config.AcmeEmail, s.config.AcmeEndpoint)
	if err != nil {
		return nil, fmt.Errorf("couldn't create certificate: %w", err)
	}

	err = s.Add(domain, cert)
	if err != nil {
		return cert, fmt.Errorf("couldn't save certificate: %w", err)
	}

	return cert, nil

}

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

	jsonData, err := json.MarshalIndent(s.certificates, "", "\t")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	err = os.WriteFile(s.config.CertificateStoreDirectory+certsFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func (s *CertificateService) loadCertificatesFromFile() error {

	jsonData, err := os.ReadFile(s.config.CertificateStoreDirectory + certsFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	err = json.Unmarshal(jsonData, &s.certificates)
	if err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return nil
}
