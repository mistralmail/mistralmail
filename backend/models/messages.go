package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Message represents an email message.
type Message struct {
	gorm.Model

	ID    uint `gorm:"primary_key;auto_increment;not_null"`
	Date  time.Time
	Size  uint32
	Flags StringSlice
	Body  []byte

	MailboxID uint `gorm:"foreignKey:Mailbox"`

	SequenceNumber uint `gorm:"->;-:migration"` // read only and skip in migrations because its the column from a view.
}

// MessageWithSequenceNumberViewName is the name of the view that represents all messages with their corresponding sequence number.
const MessageWithSequenceNumberViewName = "messages_sequence_numbers"

// MessageWithSequenceNumberViewQuery query that selects this view.
const MessageWithSequenceNumberViewQuery = `
	SELECT
		*,
		ROW_NUMBER() OVER (PARTITION BY mailbox_id ORDER BY 'date') AS sequence_number
	FROM messages
	WHERE deleted_at IS NULL;
	`

// MessageRepository implements the Message repository
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new MessageRepository
func NewMessageRepository(db *gorm.DB) (*MessageRepository, error) {
	return &MessageRepository{db: db}, nil
}

// CreateMessage creates a new message in the database.
func (r *MessageRepository) CreateMessage(message *Message) error {
	return r.db.Create(message).Error
}

// GetMessageByID retrieves a message from the database by its ID.
func (r *MessageRepository) GetMessageByID(id uint) (*Message, error) {
	var message Message
	err := r.db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// UpdateMessage updates an existing message in the database.
func (r *MessageRepository) UpdateMessage(message *Message) error {
	return r.db.Save(message).Error
}

// DeleteMessageByID deletes a message from the database by its ID.
func (r *MessageRepository) DeleteMessageByID(id uint) error {
	return r.db.Delete(&Message{}, id).Error
}

// Sequence represents a sequence of messages going from Start to Stop.
type Sequence struct {
	// Start denotes the beginning of the range (inclusive)
	Start int
	// Stop denotes the end of the range (inclusive)
	Stop int
}

// FindMessagesParameters are optional parameters that can be used to get/find messages.
type FindMessagesParameters struct {
	// SequenceSet is a list of sequences with sequence numbers.
	SequenceSet []Sequence
	// UIDSet is a list of sequences with uids.
	UIDSet []Sequence
}

// FindMessagesByMailboxID finds messages in the database by their mailbox ID.
func (r *MessageRepository) FindMessagesByMailboxID(mailboxID uint, parameters FindMessagesParameters) ([]*Message, error) {
	var messages []*Message
	query := r.db.Table(MessageWithSequenceNumberViewName).Where("mailbox_id = ?", mailboxID)

	if len(parameters.SequenceSet) > 0 && len(parameters.UIDSet) > 0 {
		return nil, fmt.Errorf("can't filter by both sequence numbers and uids at the same time")
	}

	// Handle find by sequence numbers
	var seqSetWhere *gorm.DB
	for i, sequence := range parameters.SequenceSet {

		if i == 0 {
			seqSetWhere = r.db.Where("sequence_number BETWEEN ? AND ?", sequence.Start, sequence.Stop)

		} else {
			seqSetWhere = seqSetWhere.Or("sequence_number BETWEEN ? AND ?", sequence.Start, sequence.Stop)
		}

	}
	if seqSetWhere != nil {
		query.Where(seqSetWhere)
	}

	// Handle find by uids
	var uidWhere *gorm.DB
	for i, sequence := range parameters.UIDSet {
		if i == 0 {
			uidWhere = query.Where("id BETWEEN ? AND ?", sequence.Start, sequence.Stop)

		} else {
			uidWhere = uidWhere.Or("id BETWEEN ? AND ?", sequence.Start, sequence.Stop)
		}
	}
	if uidWhere != nil {
		query.Where(uidWhere)
	}

	err := query.Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// GetNextMessageID gets the next available message id.
func (r *MessageRepository) GetNextMessageID(mailboxID uint) (uint, error) {

	message := &Message{}
	err := r.db.Order("id desc").First(&message).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 1, nil
	}
	if err != nil {
		// TODO handle error?
		return 0, fmt.Errorf("couldn't find next id: %w", err)
	}

	return message.ID + 1, nil

}

// GetNumberOfMessagesByMailboxID counts the number of messages in the given mailbox.
func (r *MessageRepository) GetNumberOfMessagesByMailboxID(mailboxID uint) (uint, error) {

	var count int64
	err := r.db.Model(&Message{}).Where(Message{MailboxID: mailboxID}).Count(&count).Error

	return uint(count), err

}

// GetTotalMessagesCount returns the total number of messages in the database.
func (r *MessageRepository) GetTotalMessagesCount() (int64, error) {
	var count int64
	if err := r.db.Model(&Message{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
