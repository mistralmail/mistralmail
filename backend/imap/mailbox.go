package imapbackend

import (
	"fmt"
	"io"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
	"github.com/mistralmail/mistralmail/backend/models"
	log "github.com/sirupsen/logrus"
)

var Delimiter = "/"

type IMAPMailbox struct {
	mailbox     *models.Mailbox
	messageRepo *models.MessageRepository
	mailboxRepo *models.MailboxRepository
}

func (mbox *IMAPMailbox) Name() string {
	return mbox.mailbox.Name
}

func (mbox *IMAPMailbox) Info() (*imap.MailboxInfo, error) {

	log.Debugln("Info")

	info := &imap.MailboxInfo{
		Delimiter: Delimiter,
		Name:      mbox.mailbox.Name,
	}
	return info, nil
}

func (mbox *IMAPMailbox) uidNext() uint {

	nextUID, err := mbox.messageRepo.GetNextMessageID(mbox.mailbox.ID)
	if err != nil {
		// TODO how to handle error?
		log.Fatalf("couldn't find next uid: %v", err)
	}

	/*
		var uid uint32
		for _, msg := range mbox.Messages {
			if msg.UID > uid {
				uid = msg.UID
			}
		}
		uid++
		return uid
	*/

	return nextUID
}

func (mbox *IMAPMailbox) flags() []string {
	/*
		flagsMap := make(map[string]bool)
		for _, msg := range mbox.Messages {
			for _, f := range msg.Flags {
				if !flagsMap[f] {
					flagsMap[f] = true
				}
			}
		}

		var flags []string
		for f := range flagsMap {
			flags = append(flags, f)
		}
		return flags
	*/

	// TODO, yep this is not performant at all, but let's hope it works
	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		// TODO handle error
		log.Fatalf("couldn't get messages: %v", err)
	}

	flagsMap := make(map[string]bool)
	for _, msg := range messages {
		for _, f := range msg.Flags {
			if !flagsMap[f] {
				flagsMap[f] = true
			}
		}
	}

	var flags []string
	for f := range flagsMap {
		flags = append(flags, f)
	}
	return flags

}

func (mbox *IMAPMailbox) unseenSeqNum() uint32 {

	// TODO, yep this is not performant at all, but let's hope it works
	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		// TODO handle error
		log.Fatalf("couldn't get messages: %v", err)
	}

	for i, msg := range messages {
		seqNum := uint32(i + 1)

		seen := false
		for _, flag := range msg.Flags {
			if flag == imap.SeenFlag {
				seen = true
				break
			}
		}

		if !seen {
			return seqNum
		}
	}
	return 0
}

func (mbox *IMAPMailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {

	log.Debugln("Status")

	count, err := mbox.messageRepo.GetNumberOfMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get status: %w", err)
	}

	status := imap.NewMailboxStatus(mbox.mailbox.Name, items)
	status.Flags = mbox.flags()
	status.PermanentFlags = []string{`\Seen`, `\Answered`, `\Flagged`, `\Draft`, `\Deleted`, `\*`}
	status.UnseenSeqNum = mbox.unseenSeqNum()

	for _, name := range items {
		switch name {
		case imap.StatusMessages:
			status.Messages = uint32(count)
		case imap.StatusUidNext:
			status.UidNext = uint32(mbox.uidNext())
		case imap.StatusUidValidity:
			status.UidValidity = 1
		case imap.StatusRecent:
			status.Recent = 0 // TODO
		case imap.StatusUnseen:
			status.Unseen = 0 // TODO
		}
	}

	return status, nil
}

func (mbox *IMAPMailbox) SetSubscribed(subscribed bool) error {
	log.Debugln("SetSubscribed")
	// TODO
	mbox.mailbox.Subscribed = subscribed

	err := mbox.mailboxRepo.UpdateMailbox(mbox.mailbox)
	if err != nil {
		return fmt.Errorf("couldn't set subscribed: %v", err)
	}

	return nil
}

func (mbox *IMAPMailbox) Check() error {
	log.Debugln("Check")
	// TODO ?
	return nil
}

func (mbox *IMAPMailbox) ListMessages(uid bool, seqSet *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	defer close(ch)

	log.Debugln("ListMessages")

	/*
		for i, msg := range mbox.Messages {
			seqNum := uint32(i + 1)

			var id uint32
			if uid {
				id = msg.UID
			} else {
				id = seqNum
			}
			if !seqSet.Contains(id) {
				continue
			}

			m, err := msg.Fetch(seqNum, items)
			if err != nil {
				continue
			}

			ch <- m
		}
	*/

	// TODO, yep this is not performant at all, but let's hope it works
	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		return fmt.Errorf("couldn't get messages: %v", err)
	}

	log.Debugf("found %d messages in mailbox %q", len(messages), mbox.Name())

	for i, message := range messages {

		msg := IMAPMessage{
			message: message,
		}

		seqNum := uint32(i + 1)

		var id uint32
		if uid {
			id = uint32(msg.message.ID)
		} else {
			id = seqNum
		}
		if !seqSet.Contains(id) {
			continue
		}

		m, err := msg.Fetch(seqNum, items)
		if err != nil {
			continue
		}

		ch <- m
	}

	return nil
}

func (mbox *IMAPMailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {

	log.Debugln("SearchMessages")

	/*
		var ids []uint32
		for i, msg := range mbox.Messages {
			seqNum := uint32(i + 1)

			ok, err := msg.Match(seqNum, criteria)
			if err != nil || !ok {
				continue
			}

			var id uint32
			if uid {
				id = msg.UID
			} else {
				id = seqNum
			}
			ids = append(ids, id)
		}
		return ids, nil
	*/

	// TODO, yep this is not performant at all, but let's hope it works
	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		return nil, fmt.Errorf("couldn't search messages: %v", err)
	}
	var ids []uint32
	for i, message := range messages {

		msg := IMAPMessage{
			message: message,
		}

		seqNum := uint32(i + 1)

		ok, err := msg.Match(seqNum, criteria)
		if err != nil || !ok {
			continue
		}

		var id uint32
		if uid {
			id = uint32(msg.message.ID)
		} else {
			id = seqNum
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (mbox *IMAPMailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {

	log.Debugln("CreateMessage")

	if date.IsZero() {
		date = time.Now()
	}

	b, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	/*
		mbox.Messages = append(mbox.Messages, &Message{
			UID:   mbox.uidNext(),
			Date:  date,
			Size:  uint32(len(b)),
			Flags: flags,
			Body:  b,
		})
	*/

	message := &models.Message{
		ID:    mbox.uidNext(),
		Date:  date,
		Size:  uint32(len(b)),
		Flags: flags,
		Body:  b,

		MailboxID: mbox.mailbox.ID,
	}

	err = mbox.messageRepo.CreateMessage(message)
	if err != nil {
		return fmt.Errorf("couldn't create message: %v", err)
	}

	return nil
}

func (mbox *IMAPMailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, op imap.FlagsOp, flags []string) error {

	log.Debugf("UpdateMessagesFlags() %+v", op)

	/*
		for i, msg := range mbox.Messages {
			var id uint32
			if uid {
				id = msg.UID
			} else {
				id = uint32(i + 1)
			}
			if !seqset.Contains(id) {
				continue
			}

			msg.Flags = backendutil.UpdateFlags(msg.Flags, op, flags)
		}
	*/

	// TODO, yep this is not performant at all, but let's hope it works
	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		return fmt.Errorf("couldn't update message flags: %v", err)
	}
	for i, message := range messages {

		msg := IMAPMessage{
			message: message,
		}

		var id uint32
		if uid {
			id = uint32(msg.message.ID)
		} else {
			id = uint32(i + 1)
		}
		if !seqset.Contains(id) {
			continue
		}

		msg.message.Flags = backendutil.UpdateFlags(msg.message.Flags, op, flags)
		err = mbox.messageRepo.UpdateMessage(msg.message)
		if err != nil {
			return fmt.Errorf("couldn't update message flags: %v", err)
		}
	}

	return nil
}

func (mbox *IMAPMailbox) CopyMessages(uid bool, seqset *imap.SeqSet, destName string) error {

	log.Debugln("CopyMessages")

	/*
		dest, ok := mbox.User.mailboxes[destName]
		if !ok {
			return backend.ErrNoSuchMailbox
		}

		for i, msg := range mbox.Messages {
			var id uint32
			if uid {
				id = msg.UID
			} else {
				id = uint32(i + 1)
			}
			if !seqset.Contains(id) {
				continue
			}

			msgCopy := *msg
			msgCopy.UID = dest.uidNext()
			dest.Messages = append(dest.Messages, &msgCopy)
		}
	*/
	destMailbox, err := mbox.mailboxRepo.GetMailBoxByUserIDAndMailboxName(mbox.mailbox.UserID, destName)
	if err != nil {
		return fmt.Errorf("couldn't find destination mailbox: %v", err)
	}
	dest := IMAPMailbox{
		mailbox:     destMailbox,
		mailboxRepo: mbox.mailboxRepo,
		messageRepo: mbox.messageRepo,
	}

	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		return fmt.Errorf("couldn't find messages: %v", err)
	}

	messagesToCopy := []*models.Message{}
	for i, msg := range messages {
		var id uint32
		if uid {
			id = uint32(msg.ID)
		} else {
			id = uint32(i + 1)
		}
		if !seqset.Contains(id) {
			continue
		}

		messagesToCopy = append(messagesToCopy, msg)
	}

	for _, message := range messagesToCopy {
		m := models.Message{
			ID:    dest.uidNext(),
			Date:  message.Date,
			Size:  message.Size,
			Flags: message.Flags,
			Body:  message.Body,

			MailboxID: dest.mailbox.ID,
		}
		err = mbox.messageRepo.CreateMessage(&m)
		if err != nil {
			return fmt.Errorf("couldn't copy messages: %v", err)
		}
	}

	return nil
}

func (mbox *IMAPMailbox) Expunge() error {

	log.Debugln("Expunge()")

	/*
		for i := len(mbox.Messages) - 1; i >= 0; i-- {
			msg := mbox.Messages[i]

			deleted := false
			for _, flag := range msg.Flags {
				if flag == imap.DeletedFlag {
					deleted = true
					break
				}
			}

			if deleted {
				mbox.Messages = append(mbox.Messages[:i], mbox.Messages[i+1:]...)
			}
		}
	*/

	// TODO, yep this is not performant at all, but let's hope it works
	messages, err := mbox.messageRepo.FindMessagesByMailboxID(mbox.mailbox.ID)
	if err != nil {
		return fmt.Errorf("couldn't find messages: %v", err)
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		deleted := false
		for _, flag := range msg.Flags {
			if flag == imap.DeletedFlag {
				deleted = true
				break
			}
		}

		if deleted {
			err := mbox.messageRepo.DeleteMessageByID(msg.ID)
			if err != nil {
				return fmt.Errorf("couldn't delete message: %v", err)
			}

		}
	}

	return nil
}
