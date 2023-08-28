package mistralmail

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given a Config instance", t, func() {

		Convey("When HOSTNAME is empty", func() {
			os.Setenv("HOSTNAME", "")
			os.Setenv("HTTP_ADDRESS", ":8080")
			os.Setenv("METRICS_ADDRESS", ":9000")
			os.Setenv("SECRET", "some-secret")
			os.Setenv("DISABLE_TLS", "true")
			os.Setenv("SMTP_ADDRESS_INCOMING", "")
			os.Setenv("SMTP_ADDRESS_OUTGOING", "smtp.outgoing.example.com:587")
			os.Setenv("SMTP_OTGOING_MODE", "RELAY")
			os.Setenv("IMAP_ADDRESS", "imap.example.com:143")
			os.Setenv("DATABASE_URL", "sqlite:file.db")

			config := BuildConfigFromEnv()

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Hostname cannot be empty")
			})
		})

		Convey("When SECRET is empty", func() {
			os.Setenv("HOSTNAME", "test")
			os.Setenv("HTTP_ADDRESS", ":8080")
			os.Setenv("METRICS_ADDRESS", ":9000")
			os.Setenv("SECRET", "")
			os.Setenv("DISABLE_TLS", "true")
			os.Setenv("SMTP_ADDRESS_INCOMING", "")
			os.Setenv("SMTP_ADDRESS_OUTGOING", "smtp.outgoing.example.com:587")
			os.Setenv("SMTP_OTGOING_MODE", "RELAY")
			os.Setenv("IMAP_ADDRESS", "imap.example.com:143")
			os.Setenv("DATABASE_URL", "sqlite:file.db")

			config := BuildConfigFromEnv()

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Secret cannot be empty")
			})
		})

		Convey("When all required fields are set", func() {
			os.Setenv("HOSTNAME", "test")
			os.Setenv("HTTP_ADDRESS", ":8080")
			os.Setenv("METRICS_ADDRESS", ":9000")
			os.Setenv("SECRET", "some-secret")
			os.Setenv("DISABLE_TLS", "true")
			os.Setenv("SMTP_ADDRESS_INCOMING", "")
			os.Setenv("SMTP_ADDRESS_OUTGOING", "smtp.outgoing.example.com:587")
			os.Setenv("SMTP_OTGOING_MODE", "RELAY")
			os.Setenv("IMAP_ADDRESS", "imap.example.com:143")
			os.Setenv("DATABASE_URL", "sqlite:file.db")

			config := BuildConfigFromEnv()

			err := config.Validate()

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})

	})
}
