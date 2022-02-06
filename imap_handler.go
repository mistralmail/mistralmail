package gopistolet

import (
	imapbackend "github.com/gopistolet/imap-backend"
	"github.com/gopistolet/smtp/mta"
	"github.com/gopistolet/smtp/smtp"
)

// NewIMAPHandler creates a new IMAP Handler
func NewIMAPHandler(c *mta.Config) *ImapHandler {
	return &ImapHandler{
		config: c,
	}
}

// ImapHandler is an SMTP handler implementation that will write mails to the IMAP backend
type ImapHandler struct {
	config *mta.Config
}

// Handle implements the SMTP Handle interface method.
// it validates the recipients email address and
// deliver the mail to the IMAP backend.
func (handler *ImapHandler) Handle(state *smtp.State) error {

	// Check whether the recipients are known to the IMAP backend
	for _, recipient := range state.To {
		recipientExists, err := imapbackend.MailaddressExists(recipient.GetAddress())
		if err != nil {
			return err
		}

		if !recipientExists {
			return smtp.SMTPErrorPermanentMailboxNotAvailable
		}
	}

	// Add mail in the backend
	_, err := imapbackend.AddMail(state)
	if err != nil {
		return err
	}

	return nil

}
