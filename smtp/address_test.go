package smtp

import (
	_ "fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseAddress(t *testing.T) {

	Convey("Testing ParseAddress()", t, func() {

		mails := []struct {
			str    string
			parsed struct {
				Local   string
				Domain  string
				Address string
			}
		}{
			{
				str: `<"bob"@example.com>`,
				parsed: struct {
					Local   string
					Domain  string
					Address string
				}{
					Local:   `bob`,
					Domain:  `example.com`,
					Address: `bob@example.com`,
				},
			},
			{
				str: `   <bob@example.com> `,
				parsed: struct {
					Local   string
					Domain  string
					Address string
				}{
					Local:   `bob`,
					Domain:  `example.com`,
					Address: `bob@example.com`,
				},
			},
			{
				str: `<" "@example.com>`,
				parsed: struct {
					Local   string
					Domain  string
					Address string
				}{
					Local:   ` `,
					Domain:  `example.com`,
					Address: ` @example.com`,
				},
			},
			{
				str: `<"test@test2"@example.com>`,
				parsed: struct {
					Local   string
					Domain  string
					Address string
				}{
					Local:   `test@test2`,
					Domain:  `example.com`,
					Address: `test@test2@example.com`,
				},
			},
		}

		for _, mail := range mails {
			address, err := ParseAddress(mail.str)
			So(err, ShouldEqual, nil)
			So(address.GetLocal(), ShouldEqual, mail.parsed.Local)
			So(address.GetDomain(), ShouldEqual, mail.parsed.Domain)
			So(address.GetAddress(), ShouldEqual, mail.parsed.Address)
		}

	})

}

/*
func TestValidate(t *testing.T) {
    Convey("Testing Validate()", t, func() {

        valid_locals := []string{
            "mathias",
            "foo,!#",
            "!def!xyz%abc",
            "$A12345",
            //"Fred Bloggs",
            "customer/department=shipping",
        }

        for _, m := range valid_locals {
            m := MailAddress{Local: m, Domain: "example.com"}
            valid, _ := m.Validate()
            So(valid, ShouldEqual, true)
        }

    })

}
*/
