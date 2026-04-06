package clock

import "time"

// SystemClock returns the current UTC time from the host system.
type SystemClock struct{}

// Now returns the current time normalized to UTC.
func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
