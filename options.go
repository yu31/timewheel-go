package timewheel

import "time"

// Option represents a modification to the default behavior of a TimeWheel.
type Option func(tw *TimeWheel)

// WithTimezone reset the timezone in TimeWheel.
func WithTimezone(loc *time.Location) Option {
	return func(tw *TimeWheel) {
		tw.location = loc
	}
}
