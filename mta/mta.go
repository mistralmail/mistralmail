package mta

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/gopistolet/gopistolet/smtp"
)

type Config struct {
	Hostname string
	Port     uint32
}

// state contains all the state for a single client
type state struct {
	from *smtp.MailAddress
	to   []*smtp.MailAddress
	data []byte
}

// reset the state
func (s *state) reset() {
	s.from = nil
	s.to = []*smtp.MailAddress{}
	s.data = []byte{}
}

// Checks the state if the client can send a MAIL command.
func (s *state) canReceiveMail() (bool, string) {
	if s.from != nil {
		return false, "Sender already specified"
	}

	return true, ""
}

// Checks the state if the client can send a RCPT command.
func (s *state) canReceiveRcpt() (bool, string) {
	if s.from == nil {
		return false, "Need mail before RCPT"
	}

	return true, ""
}

// Checks the state if the client can send a DATA command.
func (s *state) canReceiveData() (bool, string) {
	if s.from == nil {
		return false, "Need mail before DATA"
	}

	if len(s.to) == 0 {
		return false, "Need RCPT before DATA"
	}

	return true, ""
}

// Mta Represents an MTA server
type Mta struct {
	config Config
}

// New Create a new MTA server that doesn't handle the protocol.
func New(c Config) *Mta {
	mta := &Mta{
		config: c,
	}

	return mta
}

// Same as the Mta struct but has methods for handling socket connections.
type DefaultMta struct {
	mta *Mta
}

// NewDefault Create a new MTA server with a
// socket protocol implementation.
func NewDefault(c Config) *DefaultMta {
	mta := &DefaultMta{
		mta: New(c),
	}
	if mta == nil {
		return nil
	}

	return mta
}

func (s *DefaultMta) ListenAndServe() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.mta.config.Hostname, s.mta.config.Port))
	if err != nil {
		log.Printf("Could not start listening: %v", err)
		return err
	}

	return s.listen(ln)
}

func (s *DefaultMta) listen(ln net.Listener) error {
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("Accept error: %v", err)
				continue
			}
			return err
		}

		go s.serve(c)
	}

	// Dead code
	panic("Can't get here")
}

func (s *DefaultMta) serve(c net.Conn) {
	proto := smtp.NewMtaProtocol(c)
	if proto == nil {
		log.Printf("Could not create Mta protocol")
		c.Close()
		return
	}

	s.mta.HandleClient(proto)
}

// HandleClient Start communicating with a client
func (s *Mta) HandleClient(proto smtp.Protocol) {
	log.Printf("Received connection")

	// Hold state for this client connection
	state := state{}
	state.reset()

	// Start with welcome message
	proto.Send(smtp.Answer{
		Status:  smtp.Ready,
		Message: s.config.Hostname + " Service Ready",
	})

	c, ok := proto.GetCmd()
	quit := false
	for ok == true && quit == false {
		log.Printf("Received cmd: %#v", *c)

		switch cmd := (*c).(type) {
		case smtp.HeloCmd:
			proto.Send(smtp.Answer{
				Status:  smtp.Ok,
				Message: s.config.Hostname,
			})

		case smtp.QuitCmd:
			proto.Send(smtp.Answer{
				Status:  smtp.Closing,
				Message: "Bye!",
			})
			quit = true

		case smtp.MailCmd:
			if ok, reason := state.canReceiveMail(); !ok {
				proto.Send(smtp.Answer{
					Status:  smtp.BadSequence,
					Message: reason,
				})
				break
			}

			state.from = cmd.From

			proto.Send(smtp.Answer{
				Status:  smtp.Ok,
				Message: "OK",
			})

		case smtp.RcptCmd:
			if ok, reason := state.canReceiveRcpt(); !ok {
				proto.Send(smtp.Answer{
					Status:  smtp.BadSequence,
					Message: reason,
				})
				break
			}

			state.to = append(state.to, cmd.To)

			proto.Send(smtp.Answer{
				Status:  smtp.Ok,
				Message: "OK",
			})

		case smtp.DataCmd:
			if ok, reason := state.canReceiveData(); !ok {
				/*
					RFC 5321 3.3

					If there was no MAIL, or no RCPT, command, or all such commands were
					rejected, the server MAY return a "command out of sequence" (503) or
					"no valid recipients" (554) reply in response to the DATA command.
					If one of those replies (or any other 5yz reply) is received, the
					client MUST NOT send the message data; more generally, message data
					MUST NOT be sent unless a 354 reply is received.
				*/
				proto.Send(smtp.Answer{
					Status:  smtp.BadSequence,
					Message: reason,
				})
				break
			}

			proto.Send(smtp.Answer{
				Status:  smtp.StartData,
				Message: "Start mail input; end with <CRLF>.<CRLF>",
			})

		tryAgain:
			tmpData, err := ioutil.ReadAll(&cmd.R)
			state.data = append(state.data, tmpData...)
			if err == smtp.ErrLtl {
				proto.Send(smtp.Answer{
					// SyntaxError or 552 error? or something else?
					Status:  smtp.SyntaxError,
					Message: "Line too long",
				})
				goto tryAgain
			} else if err == smtp.ErrIncomplete {
				// I think this can only happen on a socket if it gets closed before receiving the full data.
				proto.Send(smtp.Answer{
					Status:  smtp.SyntaxError,
					Message: "Could not parse mail data",
				})
				state.reset()
				break

			} else if err != nil {
				panic(err)
			}

			fmt.Printf("Received mail. State: %v\n", state)

			// TODO: Handle the email

			// Reset state after mail was handled so we can start from a clean slate.
			state.reset()

		case smtp.RsetCmd:
			state.reset()
			proto.Send(smtp.Answer{
				Status:  smtp.Ok,
				Message: "OK",
			})

		case smtp.NoopCmd:
			proto.Send(smtp.Answer{
				Status:  smtp.Ok,
				Message: "OK",
			})

		case smtp.VrfyCmd, smtp.ExpnCmd, smtp.SendCmd, smtp.SomlCmd, smtp.SamlCmd:
			proto.Send(smtp.Answer{
				Status:  smtp.NotImplemented,
				Message: "Command not implemented",
			})

		case smtp.InvalidCmd:
			// TODO: Is this correct? An InvalidCmd is a known command with
			// invalid arguments. So we should send smtp.SyntaxErrorParam?
			// Is InvalidCmd a good name for this kind of error?
			proto.Send(smtp.Answer{
				Status:  smtp.SyntaxErrorParam,
				Message: cmd.Info,
			})

		case smtp.UnknownCmd:
			proto.Send(smtp.Answer{
				Status:  smtp.SyntaxError,
				Message: "Command not recognized",
			})

		default:
			// TODO: We get here if the switch does not handle all Cmd's defined
			// in protocol.go. That means we forgot to add it here. This should ideally
			// be checked at compile time. But if we get here anyway we probably shouldn't
			// crash...
			log.Fatalf("Command not implemented: %#v", cmd)
		}

		if quit {
			break
		}

		c, ok = proto.GetCmd()
	}

	proto.Close()
	log.Printf("Closed connection")
}
