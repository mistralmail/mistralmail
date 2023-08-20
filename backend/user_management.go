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

// ResetUserPassword resets the passwords of a user.
func (b *Backend) ResetUserPassword(email string, newPassword string) error {

	user, err := b.userRepo.FindUserByEmail(email)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	hashedPassword, err := models.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("couldn't hash user password: %w", err)
	}

	user.Password = hashedPassword

	err = b.userRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("couldn't save user: %w", err)
	}

	return nil
}
