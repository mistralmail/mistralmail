package smtp

import (
	_ "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParser(t *testing.T) {

	Convey("Testing parseLine()", t, func() {

		tests := []struct {
			line string
			verb string
			args []string
		}{
			{
				line: "HELO",
				verb: "HELO",
			},
			{
				line: "HELO relay.example.org",
				verb: "HELO",
				args: []string{"relay.example.org"},
			},
			{
				line: "MAIL FROM:<bob@example.org>",
				verb: "MAIL",
				args: []string{"FROM:<bob@example.org>"},
			},
		}

		for _, test := range tests {
			verb, args, err := parseLine(test.line)
			So(err, ShouldEqual, nil)
			So(verb, ShouldEqual, test.verb)
			So(args, ShouldResemble, test.args)
		}

	})

	Convey("Testing parseTo()", t, func() {

		tests := []struct {
			line          string
			addressString string
		}{
			{
				line:          "RCPT TO:<alice@example.com>",
				addressString: "alice@example.com",
			},
		}

		for _, test := range tests {
			_, args, err := parseLine(test.line)
			So(err, ShouldEqual, nil)

			addr, err := parseTO(args)
			So(err, ShouldEqual, nil)
			So(addr.GetAddress(), ShouldEqual, test.addressString)
		}

	})

}
