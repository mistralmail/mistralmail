package messageid

import (
	"net"
	"net/mail"
	"strings"
	"testing"

	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMessageIDHandler(t *testing.T) {

	Convey("Message-ID handler", t, func() {

		message := `From: sender@example.com
To: recipient@example.com
Subject: Test Subject

This is the body of the email.`

		c := server.Config{
			Hostname: "some.mail.server.example.com",
			Ip:       "192.168.0.11",
		}

		state := smtp.State{
			From:     &smtp.MailAddress{Address: "from@test.com"},
			To:       []*smtp.MailAddress{{Address: "to@test.com"}},
			Data:     []byte(message),
			Ip:       net.ParseIP("192.168.0.10"),
			Hostname: "mail.example.com",
		}

		h := New(&c)
		err := h.Handle(&state)
		So(err, ShouldEqual, nil)

		So(err, ShouldEqual, nil)

		parsedMessage, err := mail.ReadMessage(strings.NewReader(string(state.Data)))
		So(err, ShouldEqual, nil)

		So(parsedMessage.Header.Get("Message-ID"), ShouldNotBeEmpty)
		So(parsedMessage.Header.Get("Message-ID"), ShouldContainSubstring, "@some.mail.server.example.com")

	})

}
