// Copyright (c) 2020, Yu Wu <yu.771991@gmail.com> All rights reserved.
//
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package timewheel

import (
	"time"
)

type RunFunc func() error

// Schedule represents the execution plan of a task.
type Schedule interface {
	// Next returns the next execution time after the given (previous) time.
	// It will return a zero time if no next time is scheduled.
	Next(time.Time) time.Time

	// Run will be called when schedule expired.
	// Notice: timewheel will not process any errors, And only gives it to the caller.
	Run() error
}

// Schedule calls the sh.Run (in its own goroutine) according to the execution
// plan scheduled by sh.Next. It returns a Timer that can be used to cancel the
// call using its Close method.
//
// If the caller want to terminate the execution plan halfway, it must
// close the timer and wait for the timer is closed actually, since in
// the current implementation, there is a gap between the expiring and the
// restarting of the timer. The waits time is short since the gap is very small.
//
// Internally, Schedule will ask the first execution time (by calling
// sh.Next) initially, and create a timer if the execution time is non-zero.
// Afterwards, it will ask the next execution time each time task is about to
// be executed, and task will be called at the next execution time if the time
// is non-zero.
func (tw *TimeWheel) Schedule(sh Schedule) *Timer {
	next1 := sh.Next(time.Now())
	if next1.IsZero() {
		// No time is scheduled, return empty timer.
		return &Timer{}
	}
	var t *Timer
	t = &Timer{
		expiration: next1.UnixNano(),
		task: func() error {
			// Schedule the task to execute at the next time if possible.
			next2 := sh.Next(time.Unix(0, t.expiration))
			if !next2.IsZero() {
				// Resubmit the timer to next cycle.
				t.expiration = next2.UnixNano()
				tw.submit(t)
			}

			return sh.Run()
		},
		b:       nil,
		element: nil,
	}

	tw.submit(t)
	return t
}

// TimeFunc waits until the appointed time and then calls f in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Close method.
func (tw *TimeWheel) TimeFunc(t time.Time, f RunFunc) *Timer {
	return tw.expireFunc(t.UnixNano(), f)
}

// AfterFunc waits for the duration to elapse and then calls f in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Close method.
func (tw *TimeWheel) AfterFunc(d time.Duration, f RunFunc) *Timer {
	return tw.expireFunc(time.Now().Add(d).UnixNano(), f)
}

// expireFunc help creates a Timer of run-once by giving an expiration timestamp.
func (tw *TimeWheel) expireFunc(expiration int64, f RunFunc) *Timer {
	t := &Timer{
		expiration: expiration,
		task:       f,
		b:          nil,
		element:    nil,
	}

	tw.submit(t)
	return t
}
