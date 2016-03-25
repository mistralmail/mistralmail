package main

import (
	"net"

	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/smtp/smtp"

	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSaveHandler(t *testing.T) {

	Convey("Testing save() and delete() handler", t, func() {

		state := smtp.State{
			From:      &smtp.MailAddress{Address: "from@test.com"},
			To:        []*smtp.MailAddress{&smtp.MailAddress{Address: "to@test.com"}},
			Data:      []byte("Hello world!"),
			SessionId: smtp.Id{Counter: 9, Timestamp: 1455456464},
			Ip:        net.ParseIP("192.168.0.10"),
		}

		save(&state)
		stateFromFile := smtp.State{}

		filename := "mailstore/" + fileNameForState(&state)
		err := helpers.DecodeFile(filename, &stateFromFile)

		So(err, ShouldEqual, nil)
		So(state, ShouldResemble, stateFromFile)

		// Delete temp file
		delete(&state)
		So(err, ShouldEqual, nil)

	})

}
