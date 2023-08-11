package smtpbackend

import (
	"fmt"

	loginattempts "github.com/gopistolet/gopistolet/backend/login-attempts"
	"github.com/gopistolet/gopistolet/backend/models"
	"github.com/gopistolet/smtp/server"
	"github.com/gopistolet/smtp/smtp"
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
	userRepo      *models.UserRepository
	loginAttempts *loginattempts.LoginAttempts
}

// NewSMTPBackend creates a new SMTPBackend.
func NewSMTPBackend(userRepo *models.UserRepository, loginAttempts *loginattempts.LoginAttempts) (*SMTPBackend, error) {
	return &SMTPBackend{
		userRepo:      userRepo,
		loginAttempts: loginAttempts,
	}, nil
}

// Login authenticates the SMTP user.
func (b *SMTPBackend) Login(state *smtp.State, username string, password string) (server.User, error) {

	remoteIP := state.Ip.String()
	canLogin, err := b.loginAttempts.CanLogin(remoteIP)
	if err != nil {
		return nil, fmt.Errorf("couldn't check remote user: %w", err)
	}
	if !canLogin {
		return nil, fmt.Errorf("max login attempts exceeded for user with ip: %s", remoteIP)
	}

	user, err := b.userRepo.FindUserByEmail(username)
	if err != nil {
		if _, err := b.loginAttempts.AddFailedAttempts(remoteIP); err != nil {
			return nil, fmt.Errorf("couldn't increase log-in attempts: %v\n", err)
		}
		return nil, server.ErrInvalidCredentials
	}

	if user.Password == password {
		return &SMTPUser{username: username}, nil
	}

	if _, err := b.loginAttempts.AddFailedAttempts(remoteIP); err != nil {
		return nil, fmt.Errorf("couldn't increase log-in attempts: %v\n", err)
	}

	return nil, server.ErrInvalidCredentials

}
