package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {

	const (
		username          = "testuser"
		plaintextPassword = "testpassword"
		email             = "test@example.com"
	)

	user, err := NewUser(username, plaintextPassword, email)

	assert.Nil(t, err, "Error creating user should be nil")

	assert.Equal(t, username, user.Username, "Expected username should match")
	assert.NotEqual(t, plaintextPassword, user.Password, "Password should be hashed")
	assert.Equal(t, email, user.Email, "Expected email should match")
}

func TestCheckPassword(t *testing.T) {
	const (
		username          = "testuser"
		plaintextPassword = "testpassword"
		email             = "test@example.com"
	)

	user, _ := NewUser(username, plaintextPassword, email)

	match, err := user.CheckPassword(plaintextPassword)
	assert.Nil(t, err, "Error validating password should be nil")
	assert.True(t, match, "Password should match")

	match, _ = user.CheckPassword("wrongpassword")
	assert.False(t, match, "Password should not match")
}
