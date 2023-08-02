package smtpbackend

import (
	imapbackend "github.com/gopistolet/gopistolet/backend/imap"
	"github.com/gopistolet/smtp/server"
	"gorm.io/gorm"
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
	db *gorm.DB
}

// NewSMTPBackend creates a new SMTPBackend.
func NewSMTPBackend(db *gorm.DB) (*SMTPBackend, error) {

	return &SMTPBackend{db: db}, nil
}

// Login authenticates the SMTP user.
func (b *SMTPBackend) Login(username string, password string) (server.User, error) {

	user := &imapbackend.User{}

	result := b.db.Where(&imapbackend.User{Username_: username}).First(user)
	if result.RowsAffected != 1 {
		return nil, server.ErrInvalidCredentials
	}

	if user.Password == password {
		return &SMTPUser{username: username}, nil
	}

	return nil, server.ErrInvalidCredentials
}
