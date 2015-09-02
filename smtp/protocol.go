package smtp

type StatusCode uint32

// SMTP status codes
const (
	Ready            StatusCode = 220
	Closing          StatusCode = 221
	Ok               StatusCode = 250
	SyntaxError      StatusCode = 500
	SyntaxErrorParam StatusCode = 501
	NotImplemented   StatusCode = 502
	BadSequence      StatusCode = 503
)

// Cmd All SMTP answers/commands should implement this interface.
type Cmd interface{}

// Answer A raw SMTP answer. Used to send a status code + message.
type Answer struct {
	Status  StatusCode
	Message string
}

// InvalidCmd A command that is none of the other commands.
type InvalidCmd struct {
	// The command
	Cmd string
}

type HeloCmd struct {
	Domain string
}

type EhloCmd struct {
}

type QuitCmd struct {
}

type MailCmd struct {
	From *MailAddress
}

type RcptCmd struct {
	To *MailAddress
}

type DataCmd struct {
	Data []byte
}

type RsetCmd struct {
}

type NoopCmd struct{}

// Not implemented because of security concerns
type VrfyCmd struct{}
type ExpnCmd struct{}
type SendCmd struct{}
type SomlCmd struct{}
type SamlCmd struct{}

// Protocol Used as communication layer so we can easily switch between a real socket
// and a test implementation.
type Protocol interface {
	// Send a SMTP command.
	Send(Cmd)
	// Receive a command(will block while waiting for it).
	// Returns nil if no more commands(probably means the connection was closed).
	GetCmd() *Cmd
	// Close the connection.
	Close()
}
