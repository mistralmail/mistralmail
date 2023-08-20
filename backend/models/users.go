package models

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents an email user.
type User struct {
	gorm.Model

	ID       uint   `gorm:"primary_key;auto_increment;not_null"`
	Username string `gorm:"unique;not_null"`
	Password string `gorm:"not_null"`
	Email    string `gorm:"unique;not_null"`
}

// NewUser creates a new user and hashes the plaintext password.
func NewUser(username string, plaintextPassword string, email string) (*User, error) {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("couldn't hash user password: %w", err)
	}

	return &User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}, nil
}

// CheckPassword validates if the given password matches the hashed password for the user.
func (u *User) CheckPassword(plaintextPassword string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plaintextPassword))
	if err == nil {
		return true, nil
	}

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}

	return false, fmt.Errorf("couldn't validate password: %w", err)

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

	return r.db.Create(user).Error
}

// GetUserByID retrieves a user from the database by their ID.
func (r *UserRepository) GetUserByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}

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
	return &user, nil
}
