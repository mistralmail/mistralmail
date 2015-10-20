package smtp

import "net/mail"
import "strings"
import "errors"

type MailAddress mail.Address

// GetLocal gets the local part of a mail address. E.g the part before the @.
func (address *MailAddress) GetLocal() string {
	index := strings.LastIndex(address.Address, "@")
	local := address.Address[:index]
	return local
}

// GetDomain gets the domain part of a mail address. E.g the part after the @.
func (address *MailAddress) GetDomain() string {
	index := strings.LastIndex(address.Address, "@")
	domain := address.Address[index+1:]
	return domain
}

// GetAddress gets the full mail address.
func (address *MailAddress) GetAddress() string {
	return address.Address
}

// ParseAddress parses a string into a MailAddress.
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
	if index == -1 {
		return MailAddress{}, errors.New("Expected @ in mail address")
	}
	rawLocal := rawAddress[:index]
	if len(rawLocal) > 64 {
		return MailAddress{}, errors.New("Length of local part exceeds 64")
	}
	rawDomain := rawAddress[index+1:]
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
