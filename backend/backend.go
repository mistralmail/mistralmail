package backend

import (
	"fmt"

	imapbackend "github.com/mistralmail/mistralmail/backend/imap"
	loginattempts "github.com/mistralmail/mistralmail/backend/login-attempts"
	"github.com/mistralmail/mistralmail/backend/models"
	smtpbackend "github.com/mistralmail/mistralmail/backend/smtp"
	"gorm.io/gorm"
)

// Backend represents the MistralMail backend.
type Backend struct {
	db          *gorm.DB
	userRepo    *models.UserRepository
	mailboxRepo *models.MailboxRepository
	messageRepo *models.MessageRepository

	SMTPBackend *smtpbackend.SMTPBackend
	IMAPBackend *imapbackend.IMAPBackend

	loginattempts *loginattempts.LoginAttempts
}

// New creates a new backend with the provided database url.
func New(dbURL string) (*Backend, error) {

	db, err := initDB(dbURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't init db: %w", err)
	}

	userRepo, err := models.NewUserRepository(db)
	if err != nil {
		return nil, fmt.Errorf("couldn't create user repo: %w", err)
	}

	mailboxRepo, err := models.NewMailboxRepository(db)
	if err != nil {
		return nil, fmt.Errorf("couldn't create mailbox repo: %w", err)
	}

	messageRepo, err := models.NewMessageRepository(db)
	if err != nil {
		return nil, fmt.Errorf("couldn't create message repo: %w", err)
	}

	loginAttempts, err := loginattempts.New(loginattempts.DefaultMaxAttempts)
	if err != nil {
		return nil, fmt.Errorf("couldn't create login attempts service: %w", err)
	}

	imapBackend, err := imapbackend.NewIMAPBackend(userRepo, mailboxRepo, messageRepo, loginAttempts)
	if err != nil {
		return nil, fmt.Errorf("couldn't create IMAP backend: %w", err)
	}

	smtpBackend, err := smtpbackend.NewSMTPBackend(userRepo, loginAttempts)
	if err != nil {
		return nil, fmt.Errorf("couldn't create SMTP backend: %w", err)
	}

	return &Backend{
		db:            db,
		userRepo:      userRepo,
		mailboxRepo:   mailboxRepo,
		messageRepo:   messageRepo,
		IMAPBackend:   imapBackend,
		SMTPBackend:   smtpBackend,
		loginattempts: loginAttempts,
	}, nil
}

// Close the backend and its database
func (b *Backend) Close() error {
	return closeDB(b.db)
}
