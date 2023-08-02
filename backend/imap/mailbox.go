package imapbackend

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend/backendutil"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var Delimiter = "/"

type Mailbox struct {
	gorm.Model
	db *gorm.DB

	ID uint `gorm:"primary_key;auto_increment;not_null"`

	Subscribed bool
	//Messages   []*Message `gorm:"-"`

	Name_  string `gorm:"column:name;unique"`
	UserID uint   `gorm:"foreignKey:User"`
	User   *User
}

func (mbox *Mailbox) Name() string {
	return mbox.Name_
}

func (mbox *Mailbox) Info() (*imap.MailboxInfo, error) {

	log.Debugln("Info")

	info := &imap.MailboxInfo{
		Delimiter: Delimiter,
		Name:      mbox.Name_,
	}
	return info, nil
}

func (mbox *Mailbox) uidNext() uint32 {

	message := &Message{}
	err := mbox.db.Order("uid desc").First(&message).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 1
	}
	if err != nil {
		// TODO handle error?
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

	return message.UID + 1
}

func (mbox *Mailbox) flags() []string {
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
	messages := []Message{}
	err := mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
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

func (mbox *Mailbox) unseenSeqNum() uint32 {

	// TODO, yep this is not performant at all, but let's hope it works
	messages := []Message{}
	err := mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
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

func (mbox *Mailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {

	log.Debugln("Status")

	var count int64
	mbox.db.Model(&Message{}).Where(Message{MailboxID: mbox.ID}).Count(&count)

	status := imap.NewMailboxStatus(mbox.Name_, items)
	status.Flags = mbox.flags()
	status.PermanentFlags = []string{`\Seen`, `\Answered`, `\Flagged`, `\Draft`, `\Deleted`, `\*`}
	status.UnseenSeqNum = mbox.unseenSeqNum()

	for _, name := range items {
		switch name {
		case imap.StatusMessages:
			status.Messages = uint32(count)
		case imap.StatusUidNext:
			status.UidNext = mbox.uidNext()
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

func (mbox *Mailbox) SetSubscribed(subscribed bool) error {
	log.Debugln("SetSubscribed")
	// TODO
	mbox.Subscribed = subscribed

	err := mbox.db.Save(mbox).Error
	if err != nil {
		return fmt.Errorf("couldn't set subscribed: %v", err)
	}

	return nil
}

func (mbox *Mailbox) Check() error {
	log.Debugln("Check")
	return nil
}

func (mbox *Mailbox) ListMessages(uid bool, seqSet *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
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
	messages := []Message{}
	err := mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
	if err != nil {
		return fmt.Errorf("couldn't get messages: %v", err)
	}

	for i, msg := range messages {
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

	return nil
}

func (mbox *Mailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {

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
	messages := []Message{}
	err := mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("couldn't search messages: %v", err)
	}
	var ids []uint32
	for i, msg := range messages {
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
}

func (mbox *Mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {

	log.Debugln("CreateMessage")

	if date.IsZero() {
		date = time.Now()
	}

	b, err := ioutil.ReadAll(body)
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

	message := &Message{
		UID:   mbox.uidNext(),
		Date:  date,
		Size:  uint32(len(b)),
		Flags: flags,
		Body:  b,

		MailboxID: mbox.ID,
	}

	err = mbox.db.Create(message).Error
	if err != nil {
		return fmt.Errorf("couldn't create message: %v", err)
	}

	return nil
}

func (mbox *Mailbox) UpdateMessagesFlags(uid bool, seqset *imap.SeqSet, op imap.FlagsOp, flags []string) error {

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
	messages := []Message{}
	err := mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
	if err != nil {
		return fmt.Errorf("couldn't update message flags: %v", err)
	}
	for i, msg := range messages {
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
		err = mbox.db.Save(msg).Error
		if err != nil {
			return fmt.Errorf("couldn't update message flags: %v", err)
		}
	}

	return nil
}

func (mbox *Mailbox) CopyMessages(uid bool, seqset *imap.SeqSet, destName string) error {

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

	dest := &Mailbox{}
	err := mbox.db.Where(Mailbox{Name_: destName}).Find(&dest).Error
	if err != nil {
		return fmt.Errorf("couldn't find destination mailbox: %v", err)
	}

	messages := []Message{}

	err = mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
	if err != nil {
		return fmt.Errorf("couldn't find messages: %v", err)
	}

	messagesToCopy := []Message{}
	for i, msg := range messages {
		var id uint32
		if uid {
			id = msg.UID
		} else {
			id = uint32(i + 1)
		}
		if !seqset.Contains(id) {
			continue
		}

		messagesToCopy = append(messagesToCopy, msg)
	}

	for _, message := range messagesToCopy {
		m := Message{
			UID:   dest.uidNext(),
			Date:  message.Date,
			Size:  message.Size,
			Flags: message.Flags,
			Body:  message.Body,

			MailboxID: dest.ID,
		}
		err = mbox.db.Create(&m).Error
		if err != nil {
			return fmt.Errorf("couldn't copy messages: %v", err)
		}
	}

	return nil
}

func (mbox *Mailbox) Expunge() error {

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
	messages := []Message{}
	err := mbox.db.Where(Message{MailboxID: mbox.ID}).Find(&messages).Error
	if err != nil {
		return fmt.Errorf("couldn't expunge messages: %v", err)
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
			mbox.db.Delete(&msg)
		}
	}

	return nil
}
