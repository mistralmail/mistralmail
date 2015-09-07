package smtp

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type StatusCode uint32

// SMTP status codes
const (
	Ready             StatusCode = 220
	Closing           StatusCode = 221
	Ok                StatusCode = 250
	SyntaxError       StatusCode = 500
	SyntaxErrorParam  StatusCode = 501
	NotImplemented    StatusCode = 502
	BadSequence       StatusCode = 503
	NoValidRecipients StatusCode = 554
)

// Cmd All SMTP answers/commands should implement this interface.
type Cmd interface {
	fmt.Stringer
}

// Answer A raw SMTP answer. Used to send a status code + message.
type Answer struct {
	Status  StatusCode
	Message string
}

func (c Answer) String() string {
	return fmt.Sprintf("%d %s", c.Status, c.Message)
}

// InvalidCmd A command that is none of the other commands.
type InvalidCmd struct {
	// The command
	Cmd string
}

func (c InvalidCmd) String() string {
	return fmt.Sprintf("%s", c.Cmd)
}

type HeloCmd struct {
	Domain string
}

func (c HeloCmd) String() string {
	return ""
}

type EhloCmd struct {
}

func (c EhloCmd) String() string {
	return ""
}

type QuitCmd struct {
}

func (c QuitCmd) String() string {
	return ""
}

type MailCmd struct {
	From *MailAddress
}

func (c MailCmd) String() string {
	return ""
}

type RcptCmd struct {
	To *MailAddress
}

func (c RcptCmd) String() string {
	return ""
}

type DataCmd struct {
	Data []byte
}

func (c DataCmd) String() string {
	return ""
}

type RsetCmd struct {
}

func (c RsetCmd) String() string {
	return ""
}

type NoopCmd struct{}

func (c NoopCmd) String() string {
	return ""
}

// Not implemented because of security concerns
type VrfyCmd struct{}

func (c VrfyCmd) String() string {
	return ""
}

type ExpnCmd struct{}

func (c ExpnCmd) String() string {
	return ""
}

type SendCmd struct{}

func (c SendCmd) String() string {
	return ""
}

type SomlCmd struct{}

func (c SomlCmd) String() string {
	return ""
}

type SamlCmd struct{}

func (c SamlCmd) String() string {
	return ""
}

// Protocol Used as communication layer so we can easily switch between a real socket
// and a test implementation.
type Protocol interface {
	// Send a SMTP command.
	Send(Cmd)
	// Receive a command(will block while waiting for it).
	// Returns false if there are no more commands left. Otherwise a command will be returned.
	// We need the bool because if we just return nil, the nil will also implement the empty interface...
	GetCmd() (*Cmd, bool)
	// Close the connection.
	Close()
}

type MtaProtocol struct {
	c      net.Conn
	br     *bufio.Reader
	parser parser
}

// NewMtaProtocol Creates a protocol that works over a socket.
// the net.Conn parameter will be closed when done.
func NewMtaProtocol(c net.Conn) *MtaProtocol {
	proto := &MtaProtocol{
		c:      c,
		br:     bufio.NewReader(c),
		parser: parser{},
	}

	return proto
}

func (p *MtaProtocol) Send(c Cmd) {
	fmt.Fprintln(p.c, c)
}

func (p *MtaProtocol) GetCmd() (c *Cmd, ok bool) {
	cmd, err := p.parser.ParseCommand(p.br)
	if err != nil {
		log.Printf("Could not parse command: %v", err)
	}

	return &cmd, true
}

func (p *MtaProtocol) Close() {
	err := p.c.Close()
	if err != nil {
		log.Printf("Error while closing protocol: %v", err)
	}
}
