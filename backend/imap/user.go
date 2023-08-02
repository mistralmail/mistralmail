package imapbackend

import (
	"fmt"

	"github.com/emersion/go-imap/backend"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	db *gorm.DB

	ID        uint   `gorm:"primary_key;auto_increment;not_null"`
	Username_ string `gorm:"column:username;unique"`
	Password  string
	Email     string `gorm:"unique"`
	//mailboxes map[string]*Mailbox
}

func (u *User) Username() string {
	return u.Username_
}

func (u *User) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {

	mailboxes := []Mailbox{}

	err := u.db.Where(Mailbox{UserID: u.ID}).Find(&mailboxes).Error
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
		mailbox.db = u.db
		mailboxesInterface[i] = &mailbox
	}

	return mailboxesInterface, nil
}

func (u *User) GetMailbox(name string) (backend.Mailbox, error) {

	mailbox := Mailbox{db: u.db}

	err := u.db.Where(&Mailbox{Name_: name, UserID: u.ID}).First(&mailbox).Error
	if err != nil {
		return nil, fmt.Errorf("couldn't get mailbox: %w", err)
	}

	/*
		mailbox, ok := u.mailboxes[name]
		if !ok {
			err = errors.New("No such mailbox")
		}
	*/

	return &mailbox, nil
}

func (u *User) CreateMailbox(name string) error {

	/*
		if _, ok := u.mailboxes[name]; ok {
			return errors.New("Mailbox already exists")
		}

		u.mailboxes[name] = &Mailbox{Name_: name, User: u}
	*/

	mailbox := &Mailbox{
		db:     u.db,
		Name_:  name,
		User:   u,
		UserID: u.ID,
	}

	err := u.db.Create(mailbox).Error
	if err != nil {
		return fmt.Errorf("couldn't create mailbox: %v", err)
	}

	return nil
}

func (u *User) DeleteMailbox(name string) error {
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

	mailbox := &Mailbox{}
	err := u.db.Where(Mailbox{UserID: u.ID}).Where(&Mailbox{Name_: name}).Find(&mailbox).Error
	if err != nil {
		err = fmt.Errorf("mailbox does not exist: %v", err)
	}

	err = u.db.Delete(mailbox).Error
	if err != nil {
		err = fmt.Errorf("couldn't delete mailbox: %v", err)
	}

	return nil
}

func (u *User) RenameMailbox(existingName, newName string) error {

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
	mailbox := &Mailbox{}
	err := u.db.Where(Mailbox{UserID: u.ID}).Where(&Mailbox{Name_: existingName}).Find(&mailbox).Error
	if err != nil {
		err = fmt.Errorf("mailbox does not exist: %v", err)
	}

	mailbox.Name_ = newName
	err = u.db.Save(mailbox).Error
	if err != nil {
		err = fmt.Errorf("couldn't rename mailbox: %v", err)
	}

	return nil
}

func (u *User) Logout() error {
	return nil
}
