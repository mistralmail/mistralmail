package spamcheck

import (
	"errors"

	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
	log "github.com/sirupsen/logrus"
)

// New creates a new SpamCheck handler.
func New(c *server.Config) *SpamCheck {
	return &SpamCheck{
		config: c,
		api:    &PostmarkAPI{},
	}
}

// SpamCheck handler adds a Spam Score header to message.
type SpamCheck struct {
	config *server.Config
	api    SpamScoreAPI
}

// Handle gets the SpamAssassin score using Postmark SpamCheck api:
// https://spamcheck.postmarkapp.com
func (handler *SpamCheck) Handle(state *smtp.State) error {

	spamResponse, err := handler.api.getSpamScore(string(state.Data))
	if errors.Is(err, ErrEmptySpamScore) {
		// retry once
		spamResponse, err = handler.api.getSpamScore(string(state.Data))
	}
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
			"Hostname":  state.Hostname,
			"From":      state.From.String(),
			"To":        state.To,
		}).Warnf("couldn't calculate spam score: %v", err)
		return nil // don't return error here, just pass on the mail without spam score header
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
