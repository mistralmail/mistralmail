package imapbackend

import (
	"fmt"
	"time"

	"github.com/gopistolet/smtp/smtp"
)

// AddMail saves a new smtp message in the IMAP backend.
func (b *IMAPBackend) AddMail(smtpState *smtp.State) (*Message, error) {

	for _, recipient := range smtpState.To {

		user := &User{}

		err := b.DB.Where(User{Email: recipient.Address}).Find(user).Error
		if err != nil {
			return nil, fmt.Errorf("couldn't find recipient: %v", err)
		}

		inbox := &Mailbox{db: b.DB}

		err = b.DB.Where(Mailbox{UserID: user.ID, Name_: "INBOX"}).Find(inbox).Error
		if err != nil {
			return nil, fmt.Errorf("couldn't find inbox for recipient: %v", err)
		}

		message := &Message{
			UID:   inbox.uidNext(),
			Date:  time.Now(),
			Size:  uint32(len(smtpState.Data)),
			Flags: StringSlice{},
			Body:  smtpState.Data,

			MailboxID: inbox.ID,
		}

		err = b.DB.Create(message).Error
		if err != nil {
			return nil, fmt.Errorf("couldn't save new message: %v", err)
		}

	}

	return nil, nil
}

// MailaddressExists checks whether a mailbox exist for the given address.
func (b *IMAPBackend) MailaddressExists(address string) (bool, error) {

	user := &User{}
	err := b.DB.Where(User{Email: address}).Find(user).Error
	if err != nil {
		return false, fmt.Errorf("couldn't find recipient: %v", err)
	}

	if user.Email == address {
		return true, nil
	}

	return false, nil

}
