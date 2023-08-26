package backend

import (
	"fmt"

	"github.com/mistralmail/mistralmail/backend/models"
	"github.com/mistralmail/smtp/smtp"
)

// CreateNewUser creates and setups a new user with a mailbox
func (b *Backend) CreateNewUser(email string, password string) (*models.User, error) {

	// TODO validate user move somewhere else probably
	if _, err := smtp.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email address: %w", err)
	}
	if password == "" {
		// TODO have some password requirements
		return nil, fmt.Errorf("expected password")
	}

	user, err := models.NewUser(email, password, email)
	if err != nil {
		return nil, fmt.Errorf("couldn't create user: %w", err)
	}

	err = b.UserRepo.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("couldn't create user: %w", err)
	}

	mailbox := &models.Mailbox{
		Name:       "INBOX",
		Subscribed: true,
		UserID:     user.ID,
		User:       user,
	}

	err = b.MailboxRepo.CreateMailbox(mailbox)
	if err != nil {
		return nil, fmt.Errorf("couldn't create inbox for user: %w", err)
	}

	return user, nil

}

// ResetUserPassword resets the passwords of a user.
func (b *Backend) ResetUserPassword(email string, newPassword string) error {

	user, err := b.UserRepo.FindUserByEmail(email)
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	hashedPassword, err := models.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("couldn't hash user password: %w", err)
	}

	user.Password = hashedPassword

	err = b.UserRepo.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("couldn't save user: %w", err)
	}

	return nil
}
