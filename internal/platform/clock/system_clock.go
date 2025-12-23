package clock

import "time"

// SystemClock returns time.Now().
type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now()
}
