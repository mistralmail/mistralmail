package mistralmail

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/mistralmail/imap"
	"github.com/mistralmail/mistralmail/backend"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite
	client  *client.Client
	backend *backend.Backend
}

var (
	address  = "localhost:143"
	testDB   = "serve_imap_test.db"
	testUser = struct {
		Email    string
		Password string
	}{
		Email:    "test@mistralmail",
		Password: "test",
	}
)

func TestIMAPServer(t *testing.T) {

	// Create config
	config := Config{
		DatabaseURL: fmt.Sprintf("sqlite:%s", testDB),
		DisableTLS:  true,
	}

	var imapClient *client.Client

	// Create the backend
	backend, err := backend.New(config.DatabaseURL)
	if err != nil {
		t.Errorf("%v", err)
	}

	defer func() {
		imapClient.Logout()
		backend.Close()
		err := os.Remove(testDB)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()

	// Run IMAP server
	go func() {
		imapConfig := config.GenerateIMAPBackendConfig()
		imap.Serve(imapConfig, backend.IMAPBackend)
		if err != nil {
			t.Errorf("%v", err)
		}
	}()
	err = connectWithRetry(address, 5)
	if err != nil {
		t.Errorf("%v", err)
	}

	// Create a user
	_, err = backend.CreateNewUser(testUser.Email, testUser.Password)
	if err != nil {
		t.Errorf("%v", err)
	}

	Convey("Start", t, func() {

		// Init the client
		imapClient, err = client.Dial(address)
		So(err, ShouldBeNil)

		Convey("Test login with incorrect password", func() {
			username := "nonexistentuser"
			password := "wrongpassword"

			err := imapClient.Login(username, password)
			So(err, ShouldBeError)
		})

		Convey("Test login success", func() {
			err := imapClient.Login(testUser.Email, testUser.Password)
			So(err, ShouldBeNil)

			Convey("When getting the mailboxes", func() {
				mailboxes := make(chan *goimap.MailboxInfo, 10)
				done := make(chan error, 1)

				go func() {
					done <- imapClient.List("", "*", mailboxes)
				}()

				err := <-done
				So(err, ShouldBeNil)

				Convey("The mailbox 'INBOX' should exist", func() {
					exists := false
					for mailbox := range mailboxes {
						if mailbox.Name == "INBOX" {
							exists = true
							break
						}
					}
					So(exists, ShouldBeTrue)
				})
			})

			Convey("Mailboxes", func() {

				newMailbox := "NewMailbox"

				Convey("When adding a new mailbox", func() {

					err := imapClient.Create(newMailbox)
					So(err, ShouldBeNil)

					Convey("When getting the mailboxes", func() {
						mailboxes := make(chan *goimap.MailboxInfo, 10)
						done := make(chan error, 1)

						go func() {
							done <- imapClient.List("", "*", mailboxes)
						}()

						err := <-done
						So(err, ShouldBeNil)

						Convey("The new mailbox should exist", func() {
							exists := false
							for mailbox := range mailboxes {
								if mailbox.Name == newMailbox {
									exists = true
									break
								}
							}
							So(exists, ShouldBeTrue)
						})
					})

				})

				Convey("When renaming a mailbox", func() {
					oldMailbox := "OldMailbox"
					newMailbox := "AnotherNewMailbox"

					err := imapClient.Create(oldMailbox)
					So(err, ShouldBeNil)

					err = imapClient.Rename(oldMailbox, newMailbox)
					So(err, ShouldBeNil)

					// When getting the mailboxes again
					mailboxes := make(chan *goimap.MailboxInfo, 10)
					done := make(chan error, 1)

					go func() {
						done <- imapClient.List("", "*", mailboxes)
					}()

					err = <-done
					So(err, ShouldBeNil)

					// The old mailbox should not exist
					oldExists := false
					newExists := false
					for mailbox := range mailboxes {
						if mailbox.Name == oldMailbox {
							oldExists = true
						}
						if mailbox.Name == newMailbox {
							newExists = true
						}
					}
					So(oldExists, ShouldBeFalse)

					// The renamed mailbox should exist"
					So(newExists, ShouldBeTrue)

				})

				Convey("When deleting the mailbox", func() {
					err := imapClient.Delete(newMailbox)
					So(err, ShouldBeNil)

					Convey("When getting the mailboxes again", func() {
						mailboxes := make(chan *goimap.MailboxInfo, 10)
						done := make(chan error, 1)

						go func() {
							done <- imapClient.List("", "*", mailboxes)
						}()

						err := <-done
						So(err, ShouldBeNil)

						Convey("The deleted mailbox should not exist", func() {
							exists := false
							for mailbox := range mailboxes {
								if mailbox.Name == newMailbox {
									exists = true
									break
								}
							}
							So(exists, ShouldBeFalse)
						})
					})
				})

			})

		})

	})
}

func connectWithRetry(address string, maxSeconds int) error {
	timeout := time.Duration(maxSeconds) * time.Second
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("connection timed out after %d seconds", maxSeconds)
		}

		c, err := client.Dial(address)
		if err == nil {
			c.Close()
			return nil
		}

		if !isConnectionRefused(err) {
			return err
		}

		time.Sleep(1 * time.Second) // Retry after 1 second
	}
}

func isConnectionRefused(err error) bool {
	return strings.Contains(err.Error(), "refused")
}
