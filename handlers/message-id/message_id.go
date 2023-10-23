package messageid

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
	log "github.com/sirupsen/logrus"
)

// New creates a new MessageID handler with the given config.
func New(c *server.Config) *MessageID {
	return &MessageID{
		config: c,
	}
}

// MessageID handler.
type MessageID struct {
	config *server.Config
}

// Handle adds the Message-ID header to the SMTP state.
func (handler *MessageID) Handle(state *smtp.State) error {

	// Though listed as optional in the table in section 3.6, every message
	// SHOULD have a "Message-ID:" field.  Furthermore, reply messages
	// SHOULD have "In-Reply-To:" and "References:" fields as appropriate
	// and as described below.
	//
	// The "Message-ID:" field contains a single unique message identifier.
	// The "References:" and "In-Reply-To:" fields each contain one or more
	// unique message identifiers, optionally separated by CFWS.
	//
	// The message identifier (msg-id) syntax is a limited version of the
	// addr-spec construct enclosed in the angle bracket characters, "<" and
	// ">".  Unlike addr-spec, this syntax only permits the dot-atom-text
	// form on the left-hand side of the "@" and does not have internal CFWS
	// anywhere in the message identifier.

	headerKey := "Message-ID"

	headerValue := fmt.Sprintf("<%s@%s>", uuid.New(), handler.config.Hostname)

	state.AddHeader(headerKey, headerValue)

	// TODO: 'by IP' is not necessarily set in config

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
		"Hostname":  state.Hostname,
	}).Debug("Added 'Message-ID' header with value: '", headerValue, "'")

	return nil
}
