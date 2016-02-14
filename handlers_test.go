package main

import (
	"github.com/gopistolet/gopistolet/helpers"
	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"
	"net"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSaveHandler(t *testing.T) {

	Convey("Testing save() and delete() handler", t, func() {

		state := mta.State{
			From:      &smtp.MailAddress{Address: "from@test.com"},
			To:        []*smtp.MailAddress{&smtp.MailAddress{Address: "to@test.com"}},
			Data:      []byte("Hello world!"),
			SessionId: mta.Id{Counter: 9, Timestamp: 1455456464},
			Ip:        net.ParseIP("192.168.0.10"),
		}

		save(&state)
		stateFromFile := mta.State{}

		filename := "mailstore/" + fileNameForState(&state)
		err := helpers.DecodeFile(filename, &stateFromFile)

		So(err, ShouldEqual, nil)
		So(state, ShouldResemble, stateFromFile)

		// Delete temp file
		delete(&state)
		So(err, ShouldEqual, nil)

	})

}
