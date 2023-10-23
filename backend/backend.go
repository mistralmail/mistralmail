package backend

import (
	"fmt"

	imapbackend "github.com/mistralmail/mistralmail/backend/imap"
	"github.com/mistralmail/mistralmail/backend/models"
	loginattempts "github.com/mistralmail/mistralmail/backend/services/login-attempts"
	smtpbackend "github.com/mistralmail/mistralmail/backend/smtp"
	"gorm.io/gorm"
)

// Backend represents the MistralMail backend.
type Backend struct {
	db          *gorm.DB
	UserRepo    *models.UserRepository
	MailboxRepo *models.MailboxRepository
	MessageRepo *models.MessageRepository

	SMTPBackend *smtpbackend.SMTPBackend
	IMAPBackend *imapbackend.IMAPBackend

	LoginAttempts *loginattempts.LoginAttempts
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

	loginAttempts, err := loginattempts.New(loginattempts.DefaultMaxAttempts, loginattempts.DefaultBlockDuration)
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
		UserRepo:      userRepo,
		MailboxRepo:   mailboxRepo,
		MessageRepo:   messageRepo,
		IMAPBackend:   imapBackend,
		SMTPBackend:   smtpBackend,
		LoginAttempts: loginAttempts,
	}, nil
}

// Close the backend and its database
func (b *Backend) Close() error {
	return closeDB(b.db)
}
