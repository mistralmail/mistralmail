package certificates

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/lego"
	legolog "github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/registration"
	log "github.com/sirupsen/logrus"
)

// ACME is a small interface that handles the creation of certificates.
type ACME interface {
	// GenerateCertificateWithACMEChallenge generates a new certificate for the given domain.
	// the ACMEHelper must be initialized before using this function.
	GenerateCertificateWithACMEChallenge(domain string) (*CertificateResource, error)
}

// ACMEHelper is a helper struct that handles the registration of the user and the creation of certificates.
type ACMEHelper struct {
	user   acmeUser
	config *lego.Config
	client *lego.Client
}

// NewACMEHelper creates a new ACMEHelper and gets or registers the user.
func NewACMEHelper(acmePrivateKey crypto.PrivateKey, acmeEmail string, acmeEndpoint string, acmeHttpPort string, acmeHttpsPort string) (*ACMEHelper, error) {

	legolog.Logger = log.New()

	helper := &ACMEHelper{
		user: acmeUser{
			Email: acmeEmail,
			key:   acmePrivateKey,
		},
	}

	helper.config = lego.NewConfig(&helper.user)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	helper.config.CADirURL = acmeEndpoint
	helper.config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	c, err := lego.NewClient(helper.config)
	if err != nil {
		return nil, fmt.Errorf("couldn't create ACME client: %w", err)
	}
	helper.client = c

	// Setup client
	err = helper.client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", acmeHttpPort))
	if err != nil {
		return nil, fmt.Errorf("couldn't set http provider: %w", err)
	}
	err = helper.client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", acmeHttpsPort))
	if err != nil {
		return nil, fmt.Errorf("couldn't set tls provider: %w", err)
	}

	err = helper.getOrCreateUserRegistration()
	if err != nil {
		return nil, err
	}

	return helper, nil

}

// getOrCreateUserRegistration tries to get a user registration based on the private key.
// if nothing is returned it creates a new registration.
func (helper *ACMEHelper) getOrCreateUserRegistration() error {

	reg, err := helper.client.Registration.ResolveAccountByKey()
	if err != nil {
		reg, err = helper.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return fmt.Errorf("couldn't register user: %w", err)
		}
	}

	helper.user.Registration = reg

	return nil
}

// GenerateCertificateWithACMEChallenge generates a new certificate for the given domain.
// the ACMEHelper must be initialized before using this function.
func (helper *ACMEHelper) GenerateCertificateWithACMEChallenge(domain string) (*CertificateResource, error) {

	// Check if initialized
	if helper.client == nil || helper.config == nil {
		return nil, fmt.Errorf("acme helper not initialized")
	}

	// Check if user registered
	if helper.user.Registration == nil {
		return nil, fmt.Errorf("user registration not set")
	}

	// Generate certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := helper.client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("couldn't obtain certificate: %w", err)
	}

	certificateResource := &CertificateResource{
		Domain:            certificates.Domain,
		CertURL:           certificates.CertURL,
		CertStableURL:     certificates.CertStableURL,
		PrivateKey:        certificates.PrivateKey,
		Certificate:       certificates.Certificate,
		IssuerCertificate: certificates.IssuerCertificate,
		CSR:               certificates.CSR,
	}

	// TODO: this should be better and in another function...
	// way too much boilerplate to extract the certificate expiration date!
	tlsCert, err := certificateResourceToCertificate(certificateResource)
	if err != nil {
		return nil, fmt.Errorf("obtained certificate not valid: %w", err)
	}

	if len(tlsCert.Certificate) == 0 {
		return nil, fmt.Errorf("obtained certificate empty: %w", err)
	}

	var x509Cert *x509.Certificate

	for _, cert := range tlsCert.Certificate {
		x509Cert, err = x509.ParseCertificate(cert)
		if err != nil {
			return nil, fmt.Errorf("obtained sub certificate not valid: %w", err)
		}
		if x509Cert.Subject.CommonName == domain {
			break
		}
	}

	if x509Cert == nil {
		return nil, fmt.Errorf("obtained certificate not for domain: %s", err)
	}

	certificateResource.NotValidAfter = x509Cert.NotAfter

	return certificateResource, nil

}

type acmeUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *acmeUser) GetEmail() string {
	return u.Email
}
func (u acmeUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *acmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
