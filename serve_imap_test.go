package mistralmail

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/mistralmail/imap"
	"github.com/mistralmail/mistralmail/backend"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	address  = "localhost:1430"
	testDB   = "serve_imap_test.db"
	inbox    = "INBOX"
	message  = []byte("From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Test Message\r\n\r\nHello, this is a test message.\r\n")
	date     = time.Now()
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
		IMAPAddress: address,
	}

	var imapClient *client.Client

	// Create the backend
	backend, err := backend.New(config.DatabaseURL)
	if err != nil {
		t.Errorf("%v", err)
	}

	defer func() {
		err := imapClient.Logout()
		if err != nil {
			t.Logf("%v", err)
		}
		backend.Close()
		err = os.Remove(testDB)
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

				Convey("When appending a message to the inbox and getting it again and deleting one of them", func() {

					// Select mailbox
					_, err = imapClient.Select(inbox, false)
					So(err, ShouldBeNil)

					flags := []string{goimap.SeenFlag}

					// Append two messages
					err = imapClient.Append(inbox, flags, date, convertToLiteral(message))
					So(err, ShouldBeNil)

					err = imapClient.Append(inbox, flags, date, convertToLiteral(message))
					So(err, ShouldBeNil)

					// Fetch those two messages again
					seqSet := new(goimap.SeqSet)
					seqSet.AddRange(1, 2)

					messages := make(chan *goimap.Message, 2)
					done := make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchAll}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 2)

					msg := <-messages

					So(msg, ShouldNotBeNil)
					So(msg.Envelope.Subject, ShouldEqual, "Test Message")

					msg2 := <-messages

					So(msg2, ShouldNotBeNil)
					So(msg2.Envelope.Subject, ShouldEqual, "Test Message")

					// Check the number of messages in the mailbox
					status, err := imapClient.Status(inbox, []goimap.StatusItem{goimap.StatusMessages})
					So(err, ShouldBeNil)
					So(status.Messages, ShouldEqual, 2)

					// Delete the message
					seqNum := msg.SeqNum
					deletedSeqSet := new(goimap.SeqSet)
					deletedSeqSet.AddNum(seqNum)

					flags2 := []interface{}{goimap.SeenFlag, goimap.DeletedFlag}
					err = imapClient.Store(deletedSeqSet, goimap.AddFlags, flags2, nil)
					So(err, ShouldBeNil)

					// Expunge to permanently delete the message
					err = imapClient.Expunge(nil)
					So(err, ShouldBeNil)

					// Check the number of messages in the mailbox
					status, err = imapClient.Status(inbox, []goimap.StatusItem{goimap.StatusMessages})
					So(err, ShouldBeNil)
					So(status.Messages, ShouldEqual, 1)

				})

				Convey("When copying a message between mailboxes", func() {
					// Select source mailbox
					sourceMailbox := "INBOX"
					_, err := imapClient.Select(sourceMailbox, false)
					So(err, ShouldBeNil)

					// Append a message to the source mailbox
					flags := []string{goimap.SeenFlag}
					err = imapClient.Append(sourceMailbox, flags, date, convertToLiteral(message))
					So(err, ShouldBeNil)

					// Select destination mailbox
					destMailbox := "DestinationMailbox"
					err = imapClient.Create(destMailbox)
					So(err, ShouldBeNil)

					// Copy the message from source to destination
					seqSet := new(goimap.SeqSet)
					seqSet.AddNum(1) // Assuming the message is at sequence number 1
					err = imapClient.Copy(seqSet, destMailbox)
					So(err, ShouldBeNil)

					// Check that the message exists in the destination mailbox
					_, err = imapClient.Select(destMailbox, false)
					So(err, ShouldBeNil)

					seqSet = new(goimap.SeqSet)
					seqSet.AddNum(1)
					messages := make(chan *goimap.Message, 1)
					done := make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchAll}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 1)

					copiedMsg := <-messages
					So(copiedMsg, ShouldNotBeNil)
					So(copiedMsg.Envelope.Subject, ShouldEqual, "Test Message")

					// Check that the message still exists in the source mailbox
					_, err = imapClient.Select(sourceMailbox, false)
					So(err, ShouldBeNil)

					seqSet = new(goimap.SeqSet)
					seqSet.AddNum(1)
					messages = make(chan *goimap.Message, 1)
					done = make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchAll}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 1)
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

type byteLiteral struct {
	data []byte
}

func (bl *byteLiteral) Read(p []byte) (n int, err error) {
	if len(bl.data) == 0 {
		return 0, io.EOF
	}
	n = copy(p, bl.data)
	bl.data = bl.data[n:]
	return n, nil
}

func (bl *byteLiteral) Len() int {
	return len(bl.data)
}

func convertToLiteral(data []byte) goimap.Literal {
	return &byteLiteral{data: data}
}
