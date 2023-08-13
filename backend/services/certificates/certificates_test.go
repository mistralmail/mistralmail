package certificates

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testFile = "./test-certificates.json"

func TestAddAndGetCertificate(t *testing.T) {

	defer os.Remove(testFile)

	service, err := NewCertificateService(testFile, "test_endpoint", "test_email")
	require.NoError(t, err)

	certData := &CertificateResource{ /* Initialize certificate data */ }
	err = service.Add("example.com", certData)
	require.NoError(t, err)

	cert, err := service.Get("example.com")
	require.NoError(t, err)
	require.Equal(t, certData, cert)
}
