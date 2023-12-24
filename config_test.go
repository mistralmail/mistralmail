package mistralmail

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigWithoutHostname(t *testing.T) {

	Convey("When HOSTNAME is empty", t, func() {
		t.Setenv("HOSTNAME", "")
		t.Setenv("HTTP_ADDRESS", ":8080")
		t.Setenv("METRICS_ADDRESS", ":9000")
		t.Setenv("SECRET", "some-secret")
		t.Setenv("TLS_DISABLE", "true")
		t.Setenv("SMTP_ADDRESS_INCOMING", "")
		t.Setenv("SMTP_ADDRESS_OUTGOING", "smtp.outgoing.example.com:587")
		t.Setenv("SMTP_OUTGOING_MODE", "RELAY")
		t.Setenv("EXTERNAL_RELAY_HOSTNAME", "somehost")
		t.Setenv("EXTERNAL_RELAY_PORT", "587")
		t.Setenv("IMAP_ADDRESS", "imap.example.com:143")
		t.Setenv("DATABASE_URL", "sqlite:file.db")

		config, err := BuildConfigFromEnv()
		So(err, ShouldBeNil)

		err = config.Validate()

		Convey("Then it should return an error", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "HOSTNAME cannot be empty")
		})
	})

}

func TestConfigWithoutSecret(t *testing.T) {

	Convey("When SECRET is empty", t, func() {
		t.Setenv("HOSTNAME", "test")
		t.Setenv("HTTP_ADDRESS", ":8080")
		t.Setenv("METRICS_ADDRESS", ":9000")
		t.Setenv("SECRET", "")
		t.Setenv("TLS_DISABLE", "true")
		t.Setenv("SMTP_ADDRESS_INCOMING", "")
		t.Setenv("SMTP_ADDRESS_OUTGOING", "smtp.outgoing.example.com:587")
		t.Setenv("SMTP_OUTGOING_MODE", "RELAY")
		t.Setenv("EXTERNAL_RELAY_HOSTNAME", "somehost")
		t.Setenv("EXTERNAL_RELAY_PORT", "587")
		t.Setenv("IMAP_ADDRESS", "imap.example.com:143")
		t.Setenv("DATABASE_URL", "sqlite:file.db")

		config, err := BuildConfigFromEnv()
		So(err, ShouldBeNil)

		err = config.Validate()

		Convey("Then it should return an error", func() {
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "SECRET cannot be empty")
		})
	})
}

func TestConfig(t *testing.T) {

	Convey("When all required fields are set", t, func() {
		t.Setenv("HOSTNAME", "test")
		t.Setenv("HTTP_ADDRESS", ":8080")
		t.Setenv("METRICS_ADDRESS", ":9000")
		t.Setenv("SECRET", "some-secret")
		t.Setenv("TLS_DISABLE", "true")
		t.Setenv("SMTP_ADDRESS_INCOMING", "")
		t.Setenv("SMTP_ADDRESS_OUTGOING", "smtp.outgoing.example.com:587")
		t.Setenv("SMTP_OUTGOING_MODE", "RELAY")
		t.Setenv("EXTERNAL_RELAY_HOSTNAME", "somehost")
		t.Setenv("EXTERNAL_RELAY_PORT", "587")
		t.Setenv("IMAP_ADDRESS", "imap.example.com:143")
		t.Setenv("DATABASE_URL", "sqlite:file.db")

		config, err := BuildConfigFromEnv()
		So(err, ShouldBeNil)

		err = config.Validate()

		Convey("Then it should not return an error", func() {
			So(err, ShouldBeNil)
		})
	})

}
