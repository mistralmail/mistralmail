package certificates

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const testDir = "./"
const testDomain = "example.com"

func TestAddAndGetCertificate(t *testing.T) {

	defer os.Remove(testDir + certsFile)

	Convey("Add and Get certificates and save them to disk", t, func() {

		service := CertificateService{
			config: &Config{
				CertificateStoreDirectory: testDir,
			},
			certificates: map[string]*CertificateResource{},
		}

		certData := &CertificateResource{
			Domain: testDomain,
		}
		err := service.Add(testDomain, certData)
		So(err, ShouldBeNil)

		cert, err := service.Get(testDomain)
		So(err, ShouldBeNil)
		So(certData, ShouldEqual, cert)

	})

}

type mockACME struct {
}

var called = 0

func (helper *mockACME) GenerateCertificateWithACMEChallenge(domain string) (*CertificateResource, error) {
	called = called + 1
	return &CertificateResource{
		NotValidAfter: time.Now().Add(10 * time.Second),
	}, nil
}

func TestCertificateCreationAndRenewal(t *testing.T) {

	defer os.Remove(testDir + certsFile)

	Convey("Certificate creation and renewal", t, func() {

		service := CertificateService{
			config: &Config{
				CertificateStoreDirectory: testDir,
				CertificateRenewThreshold: time.Second * 1000,
			},
			certificates: map[string]*CertificateResource{},
			acmeHelper:   &mockACME{},
		}

		// Should get us a certificate
		certResource, err := service.getOrCreateCertificateResource(testDomain)
		So(err, ShouldBeNil)
		So(called, ShouldEqual, 1)
		So(certResource, ShouldNotBeNil)

		// and it should be cached
		certResource, err = service.getOrCreateCertificateResource(testDomain)
		So(err, ShouldBeNil)
		So(called, ShouldEqual, 1) // should be cached
		So(certResource, ShouldNotBeNil)

		// but also renewed if needed
		service.config.CertificateRenewThreshold = time.Second * 100 // set threshold higher than certificate validity
		service.config.CertificateRenewInterval = time.Millisecond * 10
		service.startRenewCertificateProcess()

		time.Sleep(15 * time.Millisecond)
		So(called, ShouldEqual, 2) // should be called

		_ = service

	})

}
