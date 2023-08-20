package imapbackend

import (
	"fmt"
	"net"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	loginattempts "github.com/mistralmail/mistralmail/backend/login-attempts"
	"github.com/mistralmail/mistralmail/backend/models"

	log "github.com/sirupsen/logrus"
)

type IMAPBackend struct {
	userRepo    *models.UserRepository
	mailboxRepo *models.MailboxRepository
	messageRepo *models.MessageRepository

	loginAttempts *loginattempts.LoginAttempts
}

func NewIMAPBackend(userRepo *models.UserRepository, mailboxRepo *models.MailboxRepository, messageRepo *models.MessageRepository, loginAttempts *loginattempts.LoginAttempts) (*IMAPBackend, error) {

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
		userRepo:      userRepo,
		mailboxRepo:   mailboxRepo,
		messageRepo:   messageRepo,
		loginAttempts: loginAttempts,
	}, nil
}

func (b *IMAPBackend) Login(connInfo *imap.ConnInfo, email string, password string) (backend.User, error) {

	log.WithField("remote-address", connInfo.RemoteAddr).Println("IMAP login")

	remoteIP, err := extractIPFromAddr(connInfo.RemoteAddr)
	if err != nil {
		log.WithField("remote-address", connInfo.RemoteAddr).Errorf("couldn't get ip of remote user: %v\n", err)
		return nil, fmt.Errorf("couldn't get ip of remote user: %w", err)
	}

	canLogin, err := b.loginAttempts.CanLogin(remoteIP)
	if err != nil {
		log.WithField("remote-address", connInfo.RemoteAddr).Errorf("couldn't check remote user: %v\n", err)
		return nil, fmt.Errorf("couldn't check remote user: %w", err)
	}
	if !canLogin {
		log.WithField("remote-address", connInfo.RemoteAddr).Errorf("max login attempts exceeded for user\n")
		return nil, fmt.Errorf("max login attempts exceeded for user with ip: %s", remoteIP)
	}

	user, err := b.userRepo.FindUserByEmail(email)
	if err != nil {
		if _, err := b.loginAttempts.AddFailedAttempts(remoteIP); err != nil {
			log.WithField("remote-address", connInfo.RemoteAddr).Errorf("couldn't increase log-in attempts: %v\n", err)
		}
		log.WithField("remote-address", connInfo.RemoteAddr).Errorf("bad username or password\n")
		return nil, fmt.Errorf("bad username or password")
	}

	passwordCorrect, err := user.CheckPassword(password)
	if err == nil && passwordCorrect {
		u := b.wrapUser(user)
		return &u, nil
	}
	if err != nil {
		log.WithField("remote-address", connInfo.RemoteAddr).Errorf("something went wrong checking the password: %v\n", err)
	}

	if _, err := b.loginAttempts.AddFailedAttempts(remoteIP); err != nil {
		log.WithField("remote-address", connInfo.RemoteAddr).Errorf("couldn't increase log-in attempts: %v\n", err)
	}
	log.WithField("remote-address", connInfo.RemoteAddr).Errorf("bad username or password\n")
	return nil, fmt.Errorf("bad username or password")
}

// wrapUser creates a new IMAPUser that contains the User and all repos.
func (b *IMAPBackend) wrapUser(user *models.User) IMAPUser {
	return IMAPUser{
		user:        user,
		mailboxRepo: b.mailboxRepo,
		messageRepo: b.messageRepo,
	}
}

// extractIPFromAddr extracts the IP address from a net.Addr object.
// It supports different network address types such as IP, TCP, and UDP.
// If the input address is of an unsupported type, an error is returned.
func extractIPFromAddr(addr net.Addr) (string, error) {
	switch addr := addr.(type) {
	case *net.IPAddr:
		return addr.IP.String(), nil
	case *net.TCPAddr:
		return addr.IP.String(), nil
	case *net.UDPAddr:
		return addr.IP.String(), nil
	default:
		return "", fmt.Errorf("unsupported network address type: %T", addr)
	}
}
