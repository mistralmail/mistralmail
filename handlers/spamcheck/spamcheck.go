package spamcheck

import (
	"fmt"

	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
	log "github.com/sirupsen/logrus"
)

// New creates a new SpamCheck handler.
func New(c *server.Config) *SpamCheck {
	return &SpamCheck{
		config: c,
	}
}

// SpamCheck handler adds a Spam Score header to message.
type SpamCheck struct {
	config *server.Config
}

// Handle gets the SpamAssassin score using Postmark SpamCheck api:
// https://spamcheck.postmarkapp.com
func (handler *SpamCheck) Handle(state *smtp.State) error {

	spamResponse, err := getPostMarkScore(string(state.Data))
	if err != nil {
		return fmt.Errorf("couldn't calculate spam score: %w", err)
	}

	// X-Spam-Score header
	state.AddHeader("X-Spam-Score", spamResponse.Score)

	log.WithFields(log.Fields{
		"Ip":        state.Ip.String(),
		"SessionId": state.SessionId.String(),
		"Hostname":  state.Hostname,
	}).Debugf("Spamcheck returned score of %s", spamResponse.Score)

	return nil
}
