// Copyright (c) 2020, Yu Wu <yu.771991@gmail.com> All rights reserved.
//
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package timewheel

import (
	"time"
)

// maxExpirationNs is the max nano seconds timestamp.
// In timewheel, use the UnixNano as the expiration time. Thus, the max
// expiration time is 2262-01-01 08:00:00 +0800 CST.
const maxExpirationNs = 9214646400000000000

// JobFunc is an type adapter that turns a func into an Job.
type JobFunc func() error

func (f JobFunc) Run() error {
	return f()
}

// Job used to execute job.
type Job interface {
	// Run will be called when schedule expired.
	// Notice: timewheel will not process any errors, And only gives it to the invoker.
	Run() error
}

// ScheduleFunc is a type adapter that turns a function into an Schedule.
type ScheduleFunc func(t time.Time) time.Time

func (f ScheduleFunc) Next(t time.Time) time.Time {
	return f(t)
}

// Schedule represents the execution plan of a job.
type Schedule interface {
	// Next returns the next execution time after the given (previous) time.
	// It will return a zero time if no next time is scheduled.
	Next(time.Time) time.Time
}

// ScheduleJob calls the job.Run (in its own goroutine) according to the execution
// plan scheduled by sh.Next. It returns a Timer that can be used to cancel the
// call using its Close method.
//
// If the invoker want to terminate the execution plan halfway, it must
// close the timer and wait for the timer is closed actually, since in
// the current implementation, there is a gap between the expiring and the
// restarting of the timer. The waits time is short since the gap is very small.
//
// Internally, Schedule will ask the first execution time (by calling
// sh.Next) initially, and create a timer if the execution time is non-zero.
// Afterwards, it will ask the next execution time each time task is about to
// be executed, and task will be called at the next execution time if the time
// is non-zero.
func (tw *TimeWheel) ScheduleJob(sh Schedule, job Job) *Timer {
	next1 := sh.Next(time.Now().In(tw.location))
	if next1.IsZero() {
		// No time is scheduled, return empty timer.
		return &Timer{}
	}
	expiration1 := timeToMs(next1)

	var timer *Timer
	timer = &Timer{
		expiration: expiration1,
		task: func() error {
			// ScheduleJob the task to execute at the next time if possible.
			next2 := sh.Next(msToTime(timer.expiration)).In(tw.location)
			if !next2.IsZero() {
				expiration2 := timeToMs(next2)
				// Resubmit the timer to next cycle.
				timer.expiration = expiration2
				tw.submit(timer)
			}

			return job.Run()
		},
		b:       nil,
		element: nil,
	}

	tw.submit(timer)
	return timer
}

// TimeFunc waits until the appointed time and then calls fn in its own goroutine.
// It returns a Timer that can be used to cancel the call using its Close method.
func (tw *TimeWheel) TimeFunc(t time.Time, fn JobFunc) *Timer {
	expiration1 := timeToMs(t.In(tw.location))
	timer := &Timer{
		expiration: expiration1,
		task:       fn,
		b:          nil,
		element:    nil,
	}

	tw.submit(timer)
	return timer
}
