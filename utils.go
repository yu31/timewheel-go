package timewheel

import "time"

// truncate returns the result of rounding x toward zero to a multiple of m.
// If m <= 0, Truncate returns x unchanged.
func truncate(x, m int64) int64 {
	if m <= 0 {
		return x
	}
	return x - x%m
}

func durationToMs(t time.Duration) int64 {
	return int64(t / time.Millisecond)
}

// timeToMs returns an integer number, which represents t in milliseconds.
func timeToMs(t time.Time) int64 {
	ns := t.UnixNano()
	if ns < 0 { // Means overflows int64, set it to maxExpiration
		ns = maxExpirationNs
	}
	return ns / int64(time.Millisecond)
}

// msToTime returns the UTC time corresponding to the given Unix time,
// t milliseconds since January 1, 1970 UTC.
func msToTime(t int64) time.Time {
	return time.Unix(0, t*int64(time.Millisecond))
}
