package imapbackend

import (
	"fmt"
	"time"

	"github.com/mistralmail/mistralmail/backend/models"
	"github.com/mistralmail/smtp/smtp"
)

// AddMail saves a new smtp message in the IMAP backend.
func (b *IMAPBackend) AddMail(smtpState *smtp.State) (*IMAPMessage, error) {

	for _, recipient := range smtpState.To {

		user, err := b.userRepo.FindUserByEmail(recipient.Address)
		if err != nil {
			return nil, fmt.Errorf("couldn't find recipient: %v", err)
		}

		inbox, err := b.mailboxRepo.GetMailBoxByUserIDAndMailboxName(user.ID, "INBOX")
		if err != nil {
			return nil, fmt.Errorf("couldn't find inbox for recipient: %v", err)
		}

		message := &models.Message{
			// UID:   inbox.uidNext(),
			// use gorm autoincrement in db
			Date:  time.Now(),
			Size:  uint32(len(smtpState.Data)),
			Flags: models.StringSlice{},
			Body:  smtpState.Data,

			MailboxID: inbox.ID,
		}

		err = b.messageRepo.CreateMessage(message)
		if err != nil {
			return nil, fmt.Errorf("couldn't save new message: %v", err)
		}

	}

	return nil, nil
}

// MailaddressExists checks whether a mailbox exist for the given address.
func (b *IMAPBackend) MailaddressExists(address string) (bool, error) {

	_, err := b.userRepo.FindUserByEmail(address)
	// TODO distinguis between not found and a real error.
	if err != nil {
		return false, nil
	}

	return true, nil

}
