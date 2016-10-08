package maildir

import (
	"bytes"
	"errors"

	"github.com/gopistolet/gopistolet/log"
	"github.com/gopistolet/smtp/smtp"
	"github.com/sloonz/go-maildir"
)

func New() *Maildir {
	return &Maildir{}
}

type Maildir struct {
	mailDir *maildir.Maildir
}

func (m *Maildir) Handle(state *smtp.State) {
	err := errors.New("")

	// Open maildir if it's not yet open
	if m.mailDir == nil {

		// Open a maildir. If it does not exist, create it.
		m.mailDir, err = maildir.New("./maildir", true)
		if err != nil {
			log.Errorf("Could not open maildir: %v", err)
			return
		}
	}

	dataReader := bytes.NewReader(state.Data)

	// Save mail in maildir
	filename, err := m.mailDir.CreateMail(dataReader)
	if err != nil {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Error(err)
	} else {
		log.WithFields(log.Fields{
			"Ip":        state.Ip.String(),
			"SessionId": state.SessionId.String(),
		}).Info("Maildir: mail written to file: " + filename)
	}
}
