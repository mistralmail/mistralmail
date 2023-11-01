package imapbackend

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mistralmail/mistralmail/backend/models"
	"github.com/mistralmail/smtp/smtp"
)

// AddMail saves a new smtp message in the IMAP backend.
func (b *IMAPBackend) AddMail(smtpState *smtp.State) (*IMAPMessage, error) {

	for _, recipient := range smtpState.To {

		// Find user
		user, err := b.userRepo.FindUserByEmail(recipient.Address)
		if err != nil {
			return nil, fmt.Errorf("couldn't find recipient: %w", err)
		}

		// Get either inbox or junk mailbox
		isSpam, err := isSpam(smtpState)
		if err != nil {
			return nil, fmt.Errorf("couldn't check if mail is spam: %w", err)
		}

		var destinationMailbox = &models.Mailbox{}

		if !isSpam {
			destinationMailbox, err = b.mailboxRepo.GetMailBoxByUserIDAndMailboxName(user.ID, "INBOX")
			if err != nil {
				return nil, fmt.Errorf("couldn't find inbox for recipient: %w", err)
			}
		} else {
			destinationMailbox, err = b.mailboxRepo.GetMailBoxByUserIDAndMailboxName(user.ID, "Junk")
			if err != nil {
				return nil, fmt.Errorf("couldn't find inbox for recipient: %w", err)
			}
		}

		message := &models.Message{
			// UID:   inbox.uidNext(),
			// use gorm autoincrement in db
			Date:  time.Now(),
			Size:  uint32(len(smtpState.Data)),
			Flags: models.StringSlice{},
			Body:  smtpState.Data,

			MailboxID: destinationMailbox.ID,
		}

		err = b.messageRepo.CreateMessage(message)
		if err != nil {
			return nil, fmt.Errorf("couldn't save new message: %w", err)
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

const spamThreshhold = 5.0

// isSpam checks whether a message is classified as spam, based on the X-Spam-Score header.
func isSpam(smtpState *smtp.State) (bool, error) {

	spamScore, ok := smtpState.GetHeader("X-Spam-Score")
	if !ok {
		return false, nil
	}

	spamScoreFloat, err := strconv.ParseFloat(spamScore, 64)
	if err != nil {
		return false, err
	}

	if spamScoreFloat > spamThreshhold {
		return true, nil
	}

	return false, nil

}
