package imapbackend

import (
	"errors"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/gopistolet/gopistolet/backend/models"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

type IMAPBackend struct {
	userRepo    *models.UserRepository
	mailboxRepo *models.MailboxRepository
	messageRepo *models.MessageRepository
}

func NewIMAPBackend(db *gorm.DB) (*IMAPBackend, error) {

	/*

		user := &User{Username_: "username", Password: "password"}

		body := "From: contact@example.org\r\n" +
			"To: contact@example.org\r\n" +
			"Subject: A little message, just for you\r\n" +
			"Date: Wed, 11 May 2016 14:31:59 +0000\r\n" +
			"Message-ID: <0000000@localhost/>\r\n" +
			"Content-Type: text/plain\r\n" +
			"\r\n" +
			"Hi there :)"

		user.mailboxes = map[string]*Mailbox{
			"INBOX": {
				Name_: "INBOX",
				User:  user,
				Messages: []*Message{
					{
						UID:   6,
						Date:  time.Now(),
						Flags: []string{"\\Seen"},
						Size:  uint32(len(body)),
						Body:  []byte(body),
					},
				},
			},
		}

		return &Backend{
			users: map[string]*User{user.Username_: user},
		}, nil

	*/

	userRepo, err := models.NewUserRepository(db)
	if err != nil {
		return nil, err
	}

	mailboxRepo, err := models.NewMailboxRepository(db)
	if err != nil {
		return nil, err
	}

	messageRepo, err := models.NewMessageRepository(db)
	if err != nil {
		return nil, err
	}

	return &IMAPBackend{
		userRepo:    userRepo,
		mailboxRepo: mailboxRepo,
		messageRepo: messageRepo,
	}, nil
}

func (b *IMAPBackend) Login(connInfo *imap.ConnInfo, email string, password string) (backend.User, error) {

	log.WithField("remote-address", connInfo.RemoteAddr).Println("IMAP login")

	user, err := b.userRepo.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("bad username or password")
	}

	if user.Password == password {
		u := b.wrapUser(user)
		return &u, nil
	}

	return nil, errors.New("bad username or password")
}

// wrapUser creates a new IMAPUser that contains the User and all repos.
func (b *IMAPBackend) wrapUser(user *models.User) IMAPUser {
	return IMAPUser{
		user:        user,
		mailboxRepo: b.mailboxRepo,
		messageRepo: b.messageRepo,
	}
}
