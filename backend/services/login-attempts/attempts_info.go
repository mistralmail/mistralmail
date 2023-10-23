package loginattempts

import "time"

// attemptsInfo is a smaller helpers that contains all the info for an IP address.
type attemptsInfo struct {
	ip               string
	numberOfAttempts uint
	lastAttempt      time.Time
}

// newAttemptsInfo creates a new object.
func newAttemptsInfo(ipAddress string) *attemptsInfo {
	return &attemptsInfo{
		ip:               ipAddress,
		numberOfAttempts: 1,
		lastAttempt:      time.Now(),
	}
}

// increment adds one attempts and resets the timestamp.
func (a *attemptsInfo) increment() {
	a.numberOfAttempts = a.numberOfAttempts + 1
	a.lastAttempt = time.Now()
}

// reset the number of attempts and the timestamp.
func (a *attemptsInfo) reset() {
	a.numberOfAttempts = 0
	a.lastAttempt = time.Now()
}
