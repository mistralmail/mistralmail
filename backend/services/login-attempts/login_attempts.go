package loginattempts

import "time"

// DefaultMaxAttempts denotes the default max attempts after which to block.
const DefaultMaxAttempts uint = 5

// DefaultBlockDuration denotes the default time-out during which ip addresses are blocked.
const DefaultBlockDuration = 2 * time.Hour

// LoginAttempts is a small service that counts/checks the number of failed log-in attempts.
type LoginAttempts struct {
	attemptsPerAddress map[string]*attemptsInfo
	maxAttempts        uint
	blockDuration      time.Duration
}

// New creates a new LoginAttempts service.
func New(maxAttempts uint, blockDuration time.Duration) (*LoginAttempts, error) {
	return &LoginAttempts{
		attemptsPerAddress: map[string]*attemptsInfo{},
		maxAttempts:        maxAttempts,
		blockDuration:      blockDuration,
	}, nil
}

// CanLogin checks whether the given ip address is allowed to login or blocked.
func (l *LoginAttempts) CanLogin(ipAddress string) (bool, error) {

	attempts, ok := l.attemptsPerAddress[ipAddress]
	if !ok {
		return true, nil
	}

	if time.Since(attempts.lastAttempt) > l.blockDuration {
		attempts.reset()
		return true, nil
	}

	if attempts.numberOfAttempts < l.maxAttempts {
		return true, nil
	}

	return false, nil
}

// AddFailedAttempts adds a failed login attempt for the given ip address.
func (l *LoginAttempts) AddFailedAttempts(ipAddress string) (uint, error) {

	attempts, ok := l.attemptsPerAddress[ipAddress]
	if !ok {
		l.attemptsPerAddress[ipAddress] = newAttemptsInfo(ipAddress)
	} else {
		attempts.increment()
	}

	return l.attemptsPerAddress[ipAddress].numberOfAttempts, nil
}
