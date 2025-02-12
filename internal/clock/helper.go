package clock

import "time"

// MustParse parses a formatted string and returns the time value it represents.
// The string passed should be in `RFC3339Nano` format, otherwise the function panics.
func MustParse(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, timeString)
	if err != nil {
		panic(err)
	}
	return t
}
