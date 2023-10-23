package smtpbackend

import (
	"fmt"

	"github.com/mistralmail/mistralmail/backend/models"
	loginattempts "github.com/mistralmail/mistralmail/backend/services/login-attempts"
	"github.com/mistralmail/smtp/server"
	"github.com/mistralmail/smtp/smtp"
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
			return nil, fmt.Errorf("couldn't increase log-in attempts: %w\n", err)
		}
		return nil, server.ErrInvalidCredentials
	}

	passwordCorrect, err := user.CheckPassword(password)
	if err == nil && passwordCorrect {
		return &SMTPUser{username: username}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("something went wrong checking the password: %w\n", err)
	}

	if _, err := b.loginAttempts.AddFailedAttempts(remoteIP); err != nil {
		return nil, fmt.Errorf("couldn't increase log-in attempts: %w\n", err)
	}

	return nil, server.ErrInvalidCredentials

}
