package main

import (
	_ "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseAddress(t *testing.T) {

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

}
