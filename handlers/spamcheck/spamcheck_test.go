package spamcheck

import (
	"fmt"
	"net"
	"testing"

	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSpamCheckHandler(t *testing.T) {

	Convey("Testing headerSpamCheck() handler", t, func() {

		c := server.Config{
			Hostname: "some.mail.server.example.com",
			Ip:       "192.168.0.11",
		}

		state := smtp.State{
			From:     &smtp.MailAddress{Address: "from@test.com"},
			To:       []*smtp.MailAddress{{Address: "to@test.com"}},
			Data:     []byte("Hello world!"),
			Ip:       net.ParseIP("192.168.0.10"),
			Hostname: "mail.example.com",
		}

		h := New(&c)

		// Handle with error
		h.api = mockAPIError("some error")
		err := h.Handle(&state)
		So(err, ShouldEqual, nil)

		_, ok := state.GetHeader("X-Spam-Score")
		So(ok, ShouldBeFalse)

		// Handle with correct score
		h.api = mockAPI("-5.5")
		err = h.Handle(&state)
		So(err, ShouldEqual, nil)

		header, ok := state.GetHeader("X-Spam-Score")
		So(ok, ShouldBeTrue)
		So(header, ShouldEqual, "-5.5")

	})

}

type mockAPI string

func (api mockAPI) getSpamScore(message string) (*response, error) {
	return &response{Score: string(api)}, nil
}

type mockAPIError string

func (api mockAPIError) getSpamScore(message string) (*response, error) {
	return nil, fmt.Errorf(string(api))
}
