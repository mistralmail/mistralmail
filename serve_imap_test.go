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
		err = backend.Close()
		if err != nil {
			t.Errorf("%v", err)
		}
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

				Convey("When modifying message flags", func() {
					// Select mailbox
					selectedMailbox := "INBOX"
					_, err := imapClient.Select(selectedMailbox, false)
					So(err, ShouldBeNil)

					// Append a message to the mailbox
					flags := []string{goimap.SeenFlag}
					err = imapClient.Append(selectedMailbox, flags, date, convertToLiteral(message))
					So(err, ShouldBeNil)

					// Fetch the message to get its initial flags
					seqSet := new(goimap.SeqSet)
					seqSet.AddNum(3) // Assuming the message is at sequence number 3

					messages := make(chan *goimap.Message, 1)
					done := make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchFlags}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 1)

					initialMsg := <-messages
					So(initialMsg, ShouldNotBeNil)
					So(initialMsg.Flags, ShouldContain, goimap.SeenFlag)

					// Modify the flags of the message
					newFlags := []interface{}{goimap.AnsweredFlag}
					err = imapClient.Store(seqSet, goimap.FormatFlagsOp(goimap.SetFlags, false), newFlags, nil)
					So(err, ShouldBeNil)

					// Fetch the message again to verify the updated flags
					messages = make(chan *goimap.Message, 1)
					done = make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchFlags}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 1)

					updatedMsg := <-messages
					So(updatedMsg, ShouldNotBeNil)
					So(updatedMsg.Flags, ShouldContain, goimap.AnsweredFlag)
					So(updatedMsg.Flags, ShouldNotContain, goimap.SeenFlag)
				})

			})

			Convey("Sequence Sets", func() {

				Convey("When working with sequence sets in a new mailbox", func() {
					// Create a new mailbox
					newMailbox := "NewMailboxForSequenceSets"
					err := imapClient.Create(newMailbox)
					So(err, ShouldBeNil)

					// Select the new mailbox
					_, err = imapClient.Select(newMailbox, false)
					So(err, ShouldBeNil)

					// Append three messages to the new mailbox
					for i := 1; i <= 3; i++ {
						flags := []string{goimap.SeenFlag}
						err := imapClient.Append(newMailbox, flags, date, convertToLiteral(message))
						So(err, ShouldBeNil)
					}

					// Fetching messages with a sequence set
					// Fetch messages 1 and 3
					seqSet := new(goimap.SeqSet)
					seqSet.AddRange(1, 1)
					seqSet.AddRange(3, 3)

					messages := make(chan *goimap.Message, 2)
					done := make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchUid, goimap.FetchFlags}, messages)
					}()

					select {
					case err = <-done:
						So(err, ShouldBeNil)
					case <-time.After(5 * time.Second):
						So(fmt.Errorf("time-out while waiting for fetch"), ShouldBeNil)
					}

					So(len(messages), ShouldEqual, 2)

					msg1 := <-messages
					So(msg1, ShouldNotBeNil)
					So(msg1.SeqNum, ShouldEqual, 1)

					msg3 := <-messages
					So(msg3, ShouldNotBeNil)
					So(msg3.SeqNum, ShouldEqual, 3)

					// Convey("Copying messages with a sequence set", func() {
					// Create another new mailbox for copying
					copiedMailbox := "CopiedMailboxForSequenceSets"
					err = imapClient.Create(copiedMailbox)
					So(err, ShouldBeNil)

					// Copy messages 1 and 3 to the new mailbox
					seqSet = new(goimap.SeqSet)
					seqSet.AddRange(1, 1)
					seqSet.AddRange(3, 3)

					err = imapClient.Copy(seqSet, copiedMailbox)
					So(err, ShouldBeNil)

					// Check that messages 1 and 3 exist in the copied mailbox
					_, err = imapClient.Select(copiedMailbox, false)
					So(err, ShouldBeNil)

					seqSet = new(goimap.SeqSet)
					seqSet.AddRange(1, 3)

					messages = make(chan *goimap.Message, 3)
					done = make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchUid, goimap.FetchFlags}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 2)

					msg1 = <-messages
					So(msg1, ShouldNotBeNil)
					So(msg1.SeqNum, ShouldEqual, 1)

					msg3 = <-messages
					So(msg3, ShouldNotBeNil)
					So(msg3.SeqNum, ShouldEqual, 2)

					// Deleting messages with a sequence set
					// Delete messages 1 and 3
					_, err = imapClient.Select("INBOX", false)
					So(err, ShouldBeNil)

					seqSet = new(goimap.SeqSet)
					seqSet.AddRange(1, 1)
					seqSet.AddRange(3, 3)

					err = imapClient.Store(seqSet, goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil)
					So(err, ShouldBeNil)

					// Expunge to permanently delete the messages
					err = imapClient.Expunge(nil)
					So(err, ShouldBeNil)

					// Fetch messages 1 and 3 (should not exist)
					seqSet = new(goimap.SeqSet)
					seqSet.AddRange(1, 3)

					messages = make(chan *goimap.Message, 3)
					done = make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchUid, goimap.FetchFlags}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 1)

				})
			})

			Convey("When searching for messages", func() {
				// Select the mailbox
				_, err := imapClient.Select(inbox, false)
				So(err, ShouldBeNil)

				// Append messages with different flags
				flags := []string{goimap.SeenFlag}
				err = imapClient.Append(inbox, flags, date, convertToLiteral(message))
				So(err, ShouldBeNil)

				flags = []string{goimap.AnsweredFlag}
				err = imapClient.Append(inbox, flags, date, convertToLiteral(message))
				So(err, ShouldBeNil)

				flags = []string{goimap.FlaggedFlag}
				err = imapClient.Append(inbox, flags, date, convertToLiteral(message))
				So(err, ShouldBeNil)

				// Search for messages with 'SEEN' flag
				searchCriteria := goimap.SearchCriteria{
					WithFlags: []string{"\\SEEN"},
				}

				seqNums, err := imapClient.Search(&searchCriteria)
				So(err, ShouldBeNil)

				// Fetch the matched messages
				messages := make(chan *goimap.Message, 3)
				done := make(chan error, 1)

				seqSet := &goimap.SeqSet{}
				seqSet.AddNum(seqNums...)

				go func() {
					done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchFlags}, messages)
				}()

				err = <-done
				So(err, ShouldBeNil)

				So(len(messages), ShouldEqual, 2)

				msg := <-messages
				So(msg, ShouldNotBeNil)
				So(msg.Flags, ShouldContain, goimap.SeenFlag)

				// Search for messages with 'ANSWERED' flag
				searchCriteria = goimap.SearchCriteria{
					WithFlags: []string{"\\ANSWERED"},
				}

				seqNums, err = imapClient.Search(&searchCriteria)
				So(err, ShouldBeNil)

				// Fetch the matched messages
				messages = make(chan *goimap.Message, 3)
				done = make(chan error, 1)

				seqSet = &goimap.SeqSet{}
				seqSet.AddNum(seqNums...)

				go func() {
					done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchFlags}, messages)
				}()

				err = <-done
				So(err, ShouldBeNil)

				So(len(messages), ShouldEqual, 1)

				msg = <-messages
				So(msg, ShouldNotBeNil)
				So(msg.Flags, ShouldContain, goimap.AnsweredFlag)

				// Search for messages with 'FLAGGED' flag
				searchCriteria = goimap.SearchCriteria{
					WithFlags: []string{"\\FLAGGED"},
				}

				seqNums, err = imapClient.Search(&searchCriteria)
				So(err, ShouldBeNil)

				// Fetch the matched messages
				messages = make(chan *goimap.Message, 3)
				done = make(chan error, 1)

				seqSet = &goimap.SeqSet{}
				seqSet.AddNum(seqNums...)

				go func() {
					done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchFlags}, messages)
				}()

				err = <-done
				So(err, ShouldBeNil)

				So(len(messages), ShouldEqual, 1)

				msg = <-messages
				So(msg, ShouldNotBeNil)
				So(msg.Flags, ShouldContain, goimap.FlaggedFlag)

			})

			Convey("UIDs", func() {

				Convey("When working with UIDs in a new mailbox", func() {
					// Create a new mailbox
					newMailbox := "NewMailboxForUIDs"
					err := imapClient.Create(newMailbox)
					So(err, ShouldBeNil)

					// Select the new mailbox
					_, err = imapClient.Select(newMailbox, false)
					So(err, ShouldBeNil)

					// Append three messages to the new mailbox
					for i := 1; i <= 3; i++ {
						flags := []string{goimap.SeenFlag}
						err := imapClient.Append(newMailbox, flags, date, convertToLiteral(message))
						So(err, ShouldBeNil)
					}

					// Getting the uids of the new messages
					seqSet := &goimap.SeqSet{}
					seqSet.AddNum(1, 3)

					messages := make(chan *goimap.Message, 2)
					done := make(chan error, 1)

					go func() {
						done <- imapClient.Fetch(seqSet, []goimap.FetchItem{goimap.FetchFlags, goimap.FetchUid}, messages)
					}()

					select {
					case err = <-done:
						So(err, ShouldBeNil)
					case <-time.After(5 * time.Second):
						So(fmt.Errorf("time-out while waiting for fetch"), ShouldBeNil)
					}

					So(len(messages), ShouldEqual, 2)

					msg1 := <-messages
					So(msg1, ShouldNotBeNil)
					So(msg1.Uid, ShouldNotEqual, 0)
					uid1 := msg1.Uid

					msg3 := <-messages
					So(msg3, ShouldNotBeNil)
					So(msg3.Uid, ShouldNotEqual, 0)
					uid3 := msg3.Uid

					uids := &goimap.SeqSet{}
					uids.AddNum(uid1, uid3)

					messages = make(chan *goimap.Message, 2)
					done = make(chan error, 1)

					go func() {
						done <- imapClient.UidFetch(uids, []goimap.FetchItem{goimap.FetchFlags, goimap.FetchUid}, messages)
					}()

					select {
					case err = <-done:
						So(err, ShouldBeNil)
					case <-time.After(5 * time.Second):
						So(fmt.Errorf("time-out while waiting for fetch"), ShouldBeNil)
					}

					So(len(messages), ShouldEqual, 2)

					msg1 = <-messages
					So(msg1, ShouldNotBeNil)
					So(msg1.Uid, ShouldEqual, uid1)

					msg3 = <-messages
					So(msg3, ShouldNotBeNil)
					So(msg3.Uid, ShouldEqual, uid3)

					// Deleting messages with UIDs
					_, err = imapClient.Select(newMailbox, false)
					So(err, ShouldBeNil)

					uids = &goimap.SeqSet{}
					uids.AddNum(uid1, uid3)

					err = imapClient.UidStore(uids, goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil)
					So(err, ShouldBeNil)

					// Expunge to permanently delete the messages
					err = imapClient.Expunge(nil)
					So(err, ShouldBeNil)

					// Fetch messages again
					uids = &goimap.SeqSet{}
					uids.AddNum(uid1, uid3)

					messages = make(chan *goimap.Message, 2)
					done = make(chan error, 1)

					go func() {
						done <- imapClient.UidFetch(uids, []goimap.FetchItem{goimap.FetchFlags}, messages)
					}()

					err = <-done
					So(err, ShouldBeNil)

					So(len(messages), ShouldEqual, 0)

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
