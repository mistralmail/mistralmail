package certificates

import "time"

// Resource represents a CA issued certificate. It's a copy from lego certificates.Resource
// https://pkg.go.dev/github.com/go-acme/lego/v4@v4.13.3/certificate#Resource
type CertificateResource struct {
	Domain            string
	CertURL           string
	CertStableURL     string
	PrivateKey        []byte
	Certificate       []byte
	IssuerCertificate []byte
	CSR               []byte
	NotValidAfter     time.Time
}
