package imapbackend

import (
	"fmt"

	"github.com/emersion/go-imap/backend"
	"github.com/gopistolet/gopistolet/backend/models"
)

// IMAPUser implements the emersion/go-imap User interface.
// it wraps our own user and repository into a struct.
type IMAPUser struct {
	user        *models.User
	mailboxRepo *models.MailboxRepository
	messageRepo *models.MessageRepository
}

func (u *IMAPUser) Username() string {
	return u.user.Username
}

func (u *IMAPUser) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {

	mailboxes, err := u.mailboxRepo.FindMailboxesByUserID(u.user.ID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get mailboxes: %v", err)
	}

	/*
		for _, mailbox := range u.mailboxes {
			if subscribed && !mailbox.Subscribed {
				continue
			}

			mailboxes = append(mailboxes, mailbox)
		}
	*/

	mailboxesInterface := make([]backend.Mailbox, len(mailboxes))
	for i, mailbox := range mailboxes {
		mb := u.wrapMailbox(mailbox)
		mailboxesInterface[i] = &mb
	}

	return mailboxesInterface, nil
}

func (u *IMAPUser) GetMailbox(name string) (backend.Mailbox, error) {

	mailbox, err := u.mailboxRepo.GetMailBoxByUserIDAndMailboxName(u.user.ID, name)
	if err != nil {
		return nil, fmt.Errorf("couldn't get mailbox: %w", err)
	}

	/*
		mailbox, ok := u.mailboxes[name]
		if !ok {
			err = errors.New("No such mailbox")
		}
	*/

	mb := u.wrapMailbox(mailbox)

	return &mb, nil
}

func (u *IMAPUser) CreateMailbox(name string) error {

	/*
		if _, ok := u.mailboxes[name]; ok {
			return errors.New("Mailbox already exists")
		}

		u.mailboxes[name] = &Mailbox{Name_: name, User: u}
	*/

	mailbox := &models.Mailbox{
		Name: name,
		//User:   u,
		// TODO: fixme
		UserID: u.user.ID,
	}

	err := u.mailboxRepo.CreateMailbox(mailbox)
	if err != nil {
		return fmt.Errorf("couldn't create mailbox: %v", err)
	}

	return nil
}

func (u *IMAPUser) DeleteMailbox(name string) error {
	/*
		if name == "INBOX" {
			return errors.New("Cannot delete INBOX")
		}
		if _, ok := u.mailboxes[name]; !ok {
			return errors.New("No such mailbox")
		}

		delete(u.mailboxes, name)
	*/

	if name == "INBOX" {
		return fmt.Errorf("Cannot delete INBOX")
	}

	_, err := u.mailboxRepo.GetMailBoxByUserIDAndMailboxName(u.user.ID, name)
	if err != nil {
		err = fmt.Errorf("couldn't find mailbox: %v", err)
	}

	err = u.mailboxRepo.DeleteMailboxByUserIDAndMailboxName(u.user.ID, name)
	if err != nil {
		err = fmt.Errorf("couldn't delete mailbox: %v", err)
	}

	return nil
}

func (u *IMAPUser) RenameMailbox(existingName, newName string) error {

	/*
		mbox, ok := u.mailboxes[existingName]
		if !ok {
			return errors.New("No such mailbox")
		}

		u.mailboxes[newName] = &Mailbox{
			Name_:    newName,
			Messages: mbox.Messages,
			User:     u,
		}

		mbox.Messages = nil

		if existingName != "INBOX" {
			delete(u.mailboxes, existingName)
		}
	*/
	mailbox, err := u.mailboxRepo.GetMailBoxByUserIDAndMailboxName(u.user.ID, existingName)
	if err != nil {
		err = fmt.Errorf("mailbox does not exist: %v", err)
	}

	mailbox.Name = newName
	err = u.mailboxRepo.UpdateMailbox(mailbox)
	if err != nil {
		err = fmt.Errorf("couldn't rename mailbox: %v", err)
	}

	return nil
}

func (u *IMAPUser) Logout() error {
	return nil
}

// wrapMailbox creates a new IMAMailbox that contains the Mailbox and all repos.
func (u *IMAPUser) wrapMailbox(mailbox *models.Mailbox) IMAPMailbox {
	return IMAPMailbox{
		mailbox:     mailbox,
		mailboxRepo: u.mailboxRepo,
		messageRepo: u.messageRepo,
	}
}
