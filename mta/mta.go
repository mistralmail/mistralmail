package mta

import (
	"log"

	"github.com/gopistolet/gopistolet/smtp"
)

type Config struct {
	Hostname string
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

// New Create a new MTA server
func New(c Config) *Mta {
	mta := &Mta{
		config: c,
	}

	return mta
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

	c := proto.GetCmd()
	quit := false
	for *c != nil && quit == false {
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
				proto.Send(smtp.Answer{
					Status:  smtp.BadSequence,
					Message: reason,
				})
				break
			}

			state.data = cmd.Data

			// TODO: Handle the email

			state.reset()

			proto.Send(smtp.Answer{
				Status:  smtp.Ok,
				Message: "OK",
			})

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
			proto.Send(smtp.Answer{
				Status:  smtp.SyntaxError,
				Message: "Command unrecognized",
			})

		default:
			log.Fatalf("Command not implemented: %#v", cmd)
		}

		if quit {
			break
		}

		c = proto.GetCmd()
	}

	proto.Close()
	log.Printf("Closed connection")
}
