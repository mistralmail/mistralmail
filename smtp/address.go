package smtp

import "net/mail"
import "strings"
import "errors"

type MailAddress mail.Address

func (address *MailAddress) GetLocal() string {
	index := strings.LastIndex(address.Address, "@")
	local := address.Address[0:index]
	return local
}

func (address *MailAddress) GetDomain() string {
	index := strings.LastIndex(address.Address, "@")
	domain := address.Address[index+1 : len(address.Address)]
	return domain
}

func ParseAddress(rawAddress string) (MailAddress, error) {

	/*
	   RFC 5321

	   4.5.3.1.1.  Local-part

	      The maximum total length of a user name or other local-part is 64
	      octets.

	   4.5.3.1.2.  Domain

	      The maximum total length of a domain name or number is 255 octets.
	*/
	index := strings.LastIndex(rawAddress, "@")
	rawLocal := rawAddress[0:index]
	if len(rawLocal) > 64 {
		return MailAddress{}, errors.New("Length of local part exceeds 64")
	}
	rawDomain := rawAddress[index+1 : len(rawAddress)]
	if len(rawDomain) > 255 {
		return MailAddress{}, errors.New("Length of domain name part exceeds 255")
	}

	// Try to parse the mail address using Go's built-in functions:
	address, err := mail.ParseAddress(rawAddress)
	if err != nil {
		return MailAddress{}, err
	}

	return MailAddress(*address), nil

}
