package certificates

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const testDir = "./"

func TestAddAndGetCertificate(t *testing.T) {

	defer os.Remove(testDir + certsFile)
	defer os.Remove(testDir + "/example.com.cert.pem")
	defer os.Remove(testDir + "/example.com.private.key")

	Convey("Add and Get certificates and save them to disk", t, func() {

		service := CertificateService{
			config: &Config{
				CertificateStoreDirectory: testDir,
			},
			certificates: map[string]*CertificateResource{},
		}

		certData := &CertificateResource{
			Domain: "example.com",
		}
		err := service.Add("example.com", certData)
		So(err, ShouldBeNil)

		cert, err := service.Get("example.com")
		So(err, ShouldBeNil)
		So(certData, ShouldEqual, cert)

	})

}
