package certificates

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// generateCertificateWithACMEChallenge obtains certificates using LEGO.
func generateCertificateWithACMEChallenge(domain string, acmeEmail string, acmeEndpoint string) (*CertificateResource, error) {

	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("couldn't generate private key: %w", err)
	}

	myUser := acmeUser{
		Email: acmeEmail,
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)

	// This CA URL is configured for a local dev instance of Boulder running in Docker in a VM.
	config.CADirURL = acmeEndpoint
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("couldn't create ACME client: %w", err)
	}

	// Setup client
	err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "5002"))
	if err != nil {
		return nil, fmt.Errorf("couldn't set http provider: %w", err)
	}
	err = client.Challenge.SetTLSALPN01Provider(tlsalpn01.NewProviderServer("", "5001"))
	if err != nil {
		return nil, fmt.Errorf("couldn't set tls provider: %w", err)
	}

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("couldn't set register user: %w", err)
	}
	myUser.Registration = reg

	// Generate certificate
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("couldn't obtain certificate: %w", err)
	}

	return &CertificateResource{
		Domain:            certificates.Domain,
		CertURL:           certificates.CertURL,
		CertStableURL:     certificates.CertStableURL,
		PrivateKey:        certificates.PrivateKey,
		Certificate:       certificates.Certificate,
		IssuerCertificate: certificates.IssuerCertificate,
		CSR:               certificates.CSR,
	}, nil

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
