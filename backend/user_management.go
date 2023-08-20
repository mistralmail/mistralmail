package backend

import (
	"fmt"

	"github.com/mistralmail/mistralmail/backend/models"
)

// CreateNewUser creates and setups a new user with a mailbox
func (b *Backend) CreateNewUser(email string, password string) (*models.User, error) {

	// TODO validate user

	user, err := models.NewUser(email, password, email)
	if err != nil {
		return nil, fmt.Errorf("couldn't create user: %w", err)
	}

	err = b.userRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("couldn't create user: %w", err)
	}

	mailbox := &models.Mailbox{
		Name:       "INBOX",
		Subscribed: true,
		UserID:     user.ID,
		User:       user,
	}

	err = b.mailboxRepo.CreateMailbox(mailbox)
	if err != nil {
		return nil, fmt.Errorf("couldn't create inbox for user: %w", err)
	}

	return user, nil

}
