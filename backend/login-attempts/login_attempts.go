package loginattempts

// DefaultMaxAttempts denotes the default max attempts after which to block.
const DefaultMaxAttempts uint = 5

// LoginAttempts is a small service that counts/checks the number of failed log-in attempts.
type LoginAttempts struct {
	attemptsPerAddress map[string]uint
	maxAttempts        uint
}

// New creates a new LoginAttempts service.
func New(maxAttempts uint) (*LoginAttempts, error) {
	return &LoginAttempts{
		attemptsPerAddress: map[string]uint{},
		maxAttempts:        maxAttempts,
	}, nil
}

// CanLogin checks whether the given ip address is allowed to login or blocked.
func (l *LoginAttempts) CanLogin(ipAddress string) (bool, error) {

	attempts, ok := l.attemptsPerAddress[ipAddress]
	if !ok {
		return true, nil
	}
	if attempts < l.maxAttempts {
		return true, nil
	}

	return false, nil
}

// AddFailedAttempts adds a failed login attempt for the given ip address.
func (l *LoginAttempts) AddFailedAttempts(ipAddress string) (uint, error) {

	attempts, ok := l.attemptsPerAddress[ipAddress]
	if !ok {
		l.attemptsPerAddress[ipAddress] = 1
	} else {
		l.attemptsPerAddress[ipAddress] = attempts + 1
	}

	return attempts + 1, nil
}
