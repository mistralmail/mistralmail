package models

import "gorm.io/gorm"

// User represents an email user.
type User struct {
	gorm.Model
	DB *gorm.DB `gorm:"-"`

	ID        uint   `gorm:"primary_key;auto_increment;not_null"`
	Username_ string `gorm:"column:username;unique;not_null"`
	Password  string `gorm:"not_null"`
	Email     string `gorm:"unique;not_null"`
}

// UserRepository implements the User repository
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) (*UserRepository, error) {
	return &UserRepository{db: db}, nil
}

// CreateUser creates a new user in the database.
func (r *UserRepository) CreateUser(user *User) error {

	if user.DB == nil {
		user.DB = r.db
	}

	return r.db.Create(user).Error
}

// GetUserByID retrieves a user from the database by their ID.
func (r *UserRepository) GetUserByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	user.DB = r.db

	return &user, nil
}

// UpdateUser updates an existing user in the database.
func (r *UserRepository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

// DeleteUser deletes a user from the database by their ID.
func (r *UserRepository) DeleteUser(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

// FindUserByEmail finds a user in the database by their email address.
func (r *UserRepository) FindUserByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	user.DB = r.db
	return &user, nil
}
