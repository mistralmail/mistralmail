package loginattempts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoginAttempts(t *testing.T) {
	// Test New function
	t.Run("TestNew", func(t *testing.T) {
		maxAttempts := uint(5)
		la, err := New(maxAttempts)
		assert.NoError(t, err)
		assert.NotNil(t, la)
		assert.Equal(t, maxAttempts, la.maxAttempts)
		assert.NotNil(t, la.attemptsPerAddress)
	})

	// Test CanLogin function
	t.Run("TestCanLogin", func(t *testing.T) {
		la, err := New(DefaultMaxAttempts)
		assert.NoError(t, err)

		// Test with a new IP address
		canLogin, err := la.CanLogin("192.168.0.1")
		assert.NoError(t, err)
		assert.True(t, canLogin)

		// Test with IP address exceeding max attempts
		for i := uint(0); i < DefaultMaxAttempts; i++ {
			_, _ = la.AddFailedAttempts("192.168.0.2")
		}
		canLogin, err = la.CanLogin("192.168.0.2")
		assert.NoError(t, err)
		assert.False(t, canLogin)

		// Test with IP address below max attempts
		_, _ = la.AddFailedAttempts("192.168.0.3")
		canLogin, err = la.CanLogin("192.168.0.3")
		assert.NoError(t, err)
		assert.True(t, canLogin)
	})

	// Test AddFailedAttempts function
	t.Run("TestAddFailedAttempts", func(t *testing.T) {
		la, err := New(DefaultMaxAttempts)
		assert.NoError(t, err)

		// Test adding failed attempts for a new IP address
		attempts, err := la.AddFailedAttempts("192.168.0.4")
		assert.NoError(t, err)
		assert.Equal(t, uint(1), attempts)

		// Test adding failed attempts for an existing IP address
		attempts, err = la.AddFailedAttempts("192.168.0.4")
		assert.NoError(t, err)
		assert.Equal(t, uint(2), attempts)
	})
}
