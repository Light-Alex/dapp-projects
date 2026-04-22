package utils

import "time"

// TimeToUnix converts time.Time to Unix timestamp (seconds)
func TimeToUnix(t time.Time) int64 {
	return t.Unix()
}

// UnixToTime converts Unix timestamp (seconds) to time.Time
func UnixToTime(ts int64) time.Time {
	return time.Unix(ts, 0)
}

// TimeToUnixPtr converts *time.Time to Unix timestamp (seconds)
// Returns 0 if the pointer is nil
func TimeToUnixPtr(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}

// UnixToTimePtr converts Unix timestamp (seconds) to *time.Time
// Returns nil if timestamp is 0
func UnixToTimePtr(ts int64) *time.Time {
	if ts == 0 {
		return nil
	}
	t := time.Unix(ts, 0)
	return &t
}
