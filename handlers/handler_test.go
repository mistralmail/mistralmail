package handlers

import (
	"testing"

	"github.com/mistralmail/smtp/smtp"

	. "github.com/smartystreets/goconvey/convey"
)

var count int

type TestHandler struct {
}

func (th *TestHandler) Handle(state *smtp.State) error {
	count++
	return nil
}

func TestHandlersAddress(t *testing.T) {

	// Very stupid test to make sure it does something (and keeps doing)
	Convey("Testing HandlerMechanism", t, func() {

		hm := HandlerMachanism{
			Handlers: []Handler{
				&TestHandler{},
				&TestHandler{},
			},
		}

		err := hm.Handle(nil)
		So(err, ShouldEqual, nil)

		So(count, ShouldEqual, 2)

	})

}
