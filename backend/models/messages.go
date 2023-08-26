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
}

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

// FindMessagesByMailboxID finds messages in the database by their mailbox ID.
func (r *MessageRepository) FindMessagesByMailboxID(mailboxID uint) ([]*Message, error) {
	var messages []*Message
	err := r.db.Where("mailbox_id = ?", mailboxID).Find(&messages).Error
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
