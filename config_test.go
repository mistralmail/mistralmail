package mistralmail

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given a Config instance", t, func() {
		config := &Config{}

		Convey("When SMTPAddressIncoming is empty", func() {
			config.Hostname = "test"
			config.DisableTLS = true
			config.SMTPAddressIncoming = ""
			config.SMTPAddressOutgoing = "smtp.outgoing.example.com:587"
			config.SMTPOutgoingMode = SMTPOutgoingModeRelay
			config.IMAPAddress = "imap.example.com:143"
			config.DatabaseURL = "sqlite:file.db"

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "SMTPAddressIncoming cannot be empty")
			})
		})

		Convey("When SMTPAddressOutgoing is empty", func() {
			config.Hostname = "test"
			config.DisableTLS = true
			config.SMTPAddressIncoming = "smtp.incoming.example.com:25"
			config.SMTPAddressOutgoing = ""
			config.SMTPOutgoingMode = SMTPOutgoingModeRelay
			config.IMAPAddress = "imap.example.com:143"
			config.DatabaseURL = "sqlite:file.db"

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "SMTPAddressOutgoing cannot be empty")
			})
		})

		Convey("When IMAPAddress is empty", func() {
			config.Hostname = "test"
			config.DisableTLS = true
			config.SMTPAddressIncoming = "smtp.incoming.example.com:25"
			config.SMTPAddressOutgoing = "smtp.outgoing.example.com:587"
			config.SMTPOutgoingMode = SMTPOutgoingModeRelay
			config.IMAPAddress = ""
			config.DatabaseURL = "sqlite:file.db"

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "IMAPAddress cannot be empty")
			})
		})

		Convey("When DatabaseURL is empty", func() {
			config.Hostname = "test"
			config.DisableTLS = true
			config.SMTPAddressIncoming = "smtp.incoming.example.com:25"
			config.SMTPAddressOutgoing = "smtp.outgoing.example.com:587"
			config.SMTPOutgoingMode = SMTPOutgoingModeRelay
			config.IMAPAddress = "imap.example.com:143"
			config.DatabaseURL = ""

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "DatabaseURL cannot be empty")
			})
		})

		Convey("When SMTPOutgoingMode is incorrect", func() {
			config.Hostname = "test"
			config.DisableTLS = true
			config.SMTPAddressIncoming = "smtp.incoming.example.com:25"
			config.SMTPAddressOutgoing = "smtp.outgoing.example.com:587"
			config.SMTPOutgoingMode = ""
			config.IMAPAddress = "imap.example.com:143"
			config.DatabaseURL = "sqlite:file.db"

			err := config.Validate()

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "SMTPOutgoingMode cannot be empty")
			})
		})

		Convey("When all required fields are set", func() {
			config.Hostname = "test"
			config.DisableTLS = true
			config.SMTPAddressIncoming = "smtp.incoming.example.com:25"
			config.SMTPAddressOutgoing = "smtp.outgoing.example.com:587"
			config.SMTPOutgoingMode = SMTPOutgoingModeRelay
			config.IMAPAddress = "imap.example.com:143"
			config.DatabaseURL = "sqlite:file.db"

			config.ExternalRelayHostname = "smtp.external.com"
			config.ExternalRelayPort = 587
			config.ExternalRelayUsername = "foo"
			config.ExternalRelayPassword = "bar"

			err := config.Validate()

			Convey("Then it should not return an error", func() {
				So(err, ShouldBeNil)
			})
		})

	})
}
