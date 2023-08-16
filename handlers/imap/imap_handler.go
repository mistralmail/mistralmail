package imap

import (
	imapbackend "github.com/mistralmail/mistralmail/backend/imap"
	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
)

// New creates a new IMAP Handler
func New(c *server.Config, imapbackend *imapbackend.IMAPBackend) *ImapHandler {
	return &ImapHandler{
		config:      c,
		imapbackend: imapbackend,
	}
}

// ImapHandler is an SMTP handler implementation that will write mails to the IMAP backend
type ImapHandler struct {
	config      *server.Config
	imapbackend *imapbackend.IMAPBackend
}

// Handle implements the SMTP Handle interface method.
// it validates the recipients email address and
// deliver the mail to the IMAP backend.
func (handler *ImapHandler) Handle(state *smtp.State) error {

	// Check whether the recipients are known to the IMAP backend
	for _, recipient := range state.To {
		recipientExists, err := handler.imapbackend.MailaddressExists(recipient.GetAddress())
		if err != nil {
			return err
		}

		if !recipientExists {
			return smtp.SMTPErrorPermanentMailboxNotAvailable
		}
	}

	// Add mail in the backend
	_, err := handler.imapbackend.AddMail(state)
	if err != nil {
		return err
	}

	return nil

}
