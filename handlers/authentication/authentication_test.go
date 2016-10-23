package authentication

import (
	"bytes"
	"net"
	"testing"

	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"

	. "github.com/smartystreets/goconvey/convey"
)

func TestReceivedHandler(t *testing.T) {

	Convey("Testing authenticationResultsHeader() handler", t, func() {

		/*
		   Gmail:
		       Authentication-Results: mx.google.com; spf=softfail (google.com: domain of transitioning winak@winak.be does not designate 185.27.174.242 as permitted sender) smtp.mailfrom=winak@winak.be
		*/

		c := mta.Config{
			Hostname: "some.auth.server.example.com",
			Ip:       "192.168.0.11",
		}

		state := smtp.State{
			From:     &smtp.MailAddress{Address: "from@test.com"},
			To:       []*smtp.MailAddress{&smtp.MailAddress{Address: "to@test.com"}},
			Data:     []byte("Hello world!"),
			Ip:       net.ParseIP("192.168.0.10"),
			Hostname: "mail.example.com",
		}

		h := New(&c)
		h.spfResult = "Pass"
		h.authenticationResultsHeader(&state)

		buffer := bytes.NewBuffer(state.Data)
		header, err := buffer.ReadString('\n')

		So(err, ShouldEqual, nil)
		So(header, ShouldEqual, "Authentication-Results: some.auth.server.example.com; spf=pass smtp.mailfrom=test.com;\r\n")

		h.spfResult = "SoftFail"
		h.authenticationResultsHeader(&state)

		buffer = bytes.NewBuffer(state.Data)
		header, err = buffer.ReadString('\n')

		So(err, ShouldEqual, nil)
		So(header, ShouldEqual, "Authentication-Results: some.auth.server.example.com; spf=softfail smtp.mailfrom=test.com;\r\n")

	})

}
