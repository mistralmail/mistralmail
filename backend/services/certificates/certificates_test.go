package certificates

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testDir = "./"

func TestAddAndGetCertificate(t *testing.T) {

	defer os.Remove(testDir + certsFile)
	defer os.Remove(testDir + "/example.com.cert.pem")
	defer os.Remove(testDir + "/example.com.private.key")

	service, err := NewCertificateService(testDir, "test_endpoint", "test_email")
	require.NoError(t, err)

	certData := &CertificateResource{ /* Initialize certificate data */ }
	err = service.Add("example.com", certData)
	require.NoError(t, err)

	cert, err := service.Get("example.com")
	require.NoError(t, err)
	require.Equal(t, certData, cert)
}
