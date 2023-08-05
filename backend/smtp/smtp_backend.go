package smtpbackend

import (
	"github.com/gopistolet/gopistolet/backend/models"
	"github.com/gopistolet/smtp/server"
)

// SMTPUser denotes an authenticated SMTP user.
type SMTPUser struct {
	// Username is the username / email address of the user.
	username string
}

// Username returns the username.
func (u *SMTPUser) Username() string {
	return u.username
}

// SMTPBackend implements the auth backend from the SMTP server package.
type SMTPBackend struct {
	userRepo *models.UserRepository
}

// NewSMTPBackend creates a new SMTPBackend.
func NewSMTPBackend(userRepo *models.UserRepository) (*SMTPBackend, error) {
	return &SMTPBackend{
		userRepo: userRepo,
	}, nil
}

// Login authenticates the SMTP user.
func (b *SMTPBackend) Login(username string, password string) (server.User, error) {

	user, err := b.userRepo.FindUserByEmail(username)
	if err != nil {
		return nil, server.ErrInvalidCredentials
	}

	if user.Password == password {
		return &SMTPUser{username: username}, nil
	}

	return nil, server.ErrInvalidCredentials

}
