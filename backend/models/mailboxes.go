package models

import "gorm.io/gorm"

// Mailbox represent a mailbox.
type Mailbox struct {
	gorm.Model

	ID uint `gorm:"primary_key;auto_increment;not_null"`

	Subscribed bool

	Name   string `gorm:"index:idx_mailbox_user,unique"`
	UserID uint   `gorm:"index:idx_mailbox_user,unique;foreignKey:User"`
	User   *User
}

// MailboxRepository implements the Mailbox repository
type MailboxRepository struct {
	db *gorm.DB
}

// NewMailboxRepository creates a new MailboxRepository
func NewMailboxRepository(db *gorm.DB) (*MailboxRepository, error) {
	return &MailboxRepository{db: db}, nil
}

// CreateMailbox creates a new mailbox in the database.
func (r *MailboxRepository) CreateMailbox(mailbox *Mailbox) error {
	return r.db.Create(mailbox).Error
}

// GetMailboxByID retrieves a mailbox from the database by its ID.
func (r *MailboxRepository) GetMailboxByID(id uint) (*Mailbox, error) {
	var mailbox Mailbox
	err := r.db.First(&mailbox, id).Error
	if err != nil {
		return nil, err
	}
	return &mailbox, nil
}

// UpdateMailbox updates an existing mailbox in the database.
func (r *MailboxRepository) UpdateMailbox(mailbox *Mailbox) error {
	return r.db.Save(mailbox).Error
}

// DeleteMailbox deletes a mailbox from the database by its ID.
func (r *MailboxRepository) DeleteMailbox(id uint) error {
	return r.db.Delete(&Mailbox{}, id).Error
}

// DeleteMailboxByUserIDAndMailboxName deletes a mailbox from the database by user ID and mailbox name.
func (r *MailboxRepository) DeleteMailboxByUserIDAndMailboxName(userID uint, mailboxName string) error {
	return r.db.Delete(&Mailbox{}, "user_id = ? AND name = ?", userID, mailboxName).Error
}

// FindMailboxesByUserID finds mailboxes in the database by their user ID.
func (r *MailboxRepository) FindMailboxesByUserID(userID uint) ([]*Mailbox, error) {
	var mailboxes []*Mailbox
	err := r.db.Where("user_id = ?", userID).Find(&mailboxes).Error
	if err != nil {
		return nil, err
	}
	return mailboxes, nil
}

// GetMailBoxByUserIDAndMailboxName retrieves a mailbox from the database by user ID and mailbox name.
func (r *MailboxRepository) GetMailBoxByUserIDAndMailboxName(userID uint, mailboxName string) (*Mailbox, error) {
	var mailbox Mailbox
	err := r.db.Where("user_id = ? AND name = ?", userID, mailboxName).First(&mailbox).Error
	if err != nil {
		return nil, err
	}
	return &mailbox, nil
}
