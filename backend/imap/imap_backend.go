package imapbackend

import (
	"errors"
	"fmt"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

type IMAPBackend struct {
	DB    *gorm.DB
	users map[string]*User
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
	return &IMAPBackend{
		DB: db,
	}, nil
}

func (b *IMAPBackend) Login(connInfo *imap.ConnInfo, username string, password string) (backend.User, error) {

	log.WithField("remote-address", connInfo.RemoteAddr).Println("login")

	user := &User{db: b.DB}

	result := b.DB.Where(&User{Username_: username}).First(user)
	if result.RowsAffected != 1 {
		return nil, fmt.Errorf("bad username or password")
	}

	if user.Password == password {
		return user, nil
	}

	return nil, errors.New("bad username or password")
}
