// Copyright (c) 2020, Yu Wu <yu.771991@gmail.com> All rights reserved.
//
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package timewheel

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/yu31/dqueue-go"
)

const (
	defaultTick = time.Millisecond
	defaultSize = int64(32)
)

// TimeWheel is an implementation of Hierarchical Timing Wheels.
type TimeWheel struct {
	tick    int64 // The time span of each unit, in milliseconds(nanoseconds/time.Millisecond).
	size    int64 // The size of time wheel of each layer.
	span    int64 // The time span of each layer, in milliseconds(nanoseconds/time.Millisecond).
	current int64 // The current time of time wheel, in milliseconds(nanoseconds/time.Millisecond).

	buckets []*bucket
	queue   *dqueue.DQueue

	// The time zone. default is time.Local
	location *time.Location

	// Store the options.
	opts []Option

	// The higher-level overflow TimeWheel.
	//
	// NOTICE: This field may be updated and read concurrently, through tw.add().
	overflow unsafe.Pointer // type: *TimingWheel
}

// Default creates an TimeWheel with default parameters.
func Default(opts ...Option) *TimeWheel {
	return New(defaultTick, defaultSize, opts...)
}

// New creates an TimeWheel with the given tick and wheel size.
// The value of tick must >= 1ms, the size must >= 1.
func New(tick time.Duration, size int64, opts ...Option) *TimeWheel {
	if tick < time.Millisecond {
		panic("timewheel: tick must be greater than or equal to 1ms")
	}
	if size < 1 {
		panic("timewheel: size must be greater than 0")
	}
	tickMs := durationToMs(tick)
	startMs := timeToMs(time.Now())
	dq := dqueue.Default().WithDelayer(func(expiration int64) (delay time.Duration) {
		return time.Duration(expiration - timeToMs(time.Now()))
	})

	return newTimeWheel(tickMs, size, startMs, dq, opts...)
}

// newTimeWheel is an internal helper function that really creates an TimeWheel.
func newTimeWheel(tickMs int64, size int64, startMs int64, queue *dqueue.DQueue, opts ...Option) *TimeWheel {
	tw := &TimeWheel{
		tick:     tickMs,
		size:     size,
		span:     tickMs * size,
		current:  truncate(startMs, tickMs),
		buckets:  createBuckets(int(size)),
		queue:    queue,
		location: time.Local,
		opts:     opts,
		overflow: nil,
	}
	for _, opt := range opts {
		opt(tw)
	}
	return tw
}

// Start starts the current time wheel in a goroutine.
// You can call the Wait method to blocks the main process after.
func (tw *TimeWheel) Start() {
	tw.queue.Start(tw.process)
}

// Stop stops the current time wheel.
//
// If there is any timer's jobFunc being running in its own goroutine, Stop does
// not wait for the jobFunc to complete before returning. If the invoker needs to
// know whether the jobFunc is completed, it must coordinate with the jobFunc explicitly.
func (tw *TimeWheel) Stop() {
	tw.queue.Stop()
}

// process the expiration's bucket
func (tw *TimeWheel) process(val dqueue.Value) {
	b := val.(*bucket)
	tw.advance(b.getExpiration())

	b.flush(tw.submit)
}

// advance push the clock forward.
func (tw *TimeWheel) advance(expiration int64) {
	current := atomic.LoadInt64(&tw.current)
	if expiration >= current+tw.tick {
		current = truncate(expiration, tw.tick)
		atomic.StoreInt64(&tw.current, current)

		// Try to advance the clock of the overflow wheel if present
		overflow := atomic.LoadPointer(&tw.overflow)
		if overflow != nil {
			(*TimeWheel)(overflow).advance(current)
		}
	}
}

// submit inserts the timer t into the current timing wheel, or run the
// timer's jobFunc if it has been expired.
func (tw *TimeWheel) submit(t *Timer) {
	if !tw.add(t) {
		// Actually execute the jobFunc func.
		//
		// Like the standard time.AfterFunc (https://golang.org/pkg/time/#AfterFunc),
		// always execute the timer's jobFunc in its own goroutine.
		go func() {
			_ = t.jobFunc(t.ctxCancel)
		}()
	}
}

// add inserts the timer t into the current timing wheel.
// return false means the Timer has been expired.
func (tw *TimeWheel) add(t *Timer) bool {
	current := atomic.LoadInt64(&tw.current)
	if t.expiration < current+tw.tick {
		// Already expired.
		return false
	} else if t.expiration < current+tw.span {
		// Put it into its own bucket.
		virtualId := t.expiration / tw.tick
		b := tw.buckets[virtualId%tw.size]
		b.insert(t)

		// Set the bucket expiration timestamp.
		if b.setExpiration(virtualId * tw.tick) {
			// The bucket needs to be enqueued since it was an expired bucket.
			// We only need to enqueue the bucket when its expiration time has changed,
			// i.e. the wheel has advanced and this bucket get reused with a new expiration.
			// Any further calls to set the expiration within the same wheel cycle will
			// pass in the same value and hence return false, thus the bucket with the
			// same expiration will not be enqueued multiple times.
			tw.queue.Offer(b.getExpiration(), b)
		}
		return true
	} else {
		// Out of the span. Put it into the overflow TimeWheel.
		var overflow unsafe.Pointer

		overflow = atomic.LoadPointer(&tw.overflow)
		if overflow == nil {
			// Creates and save overflow TimeWheel.
			ntw := newTimeWheel(tw.span, tw.size, current, tw.queue, tw.opts...)
			atomic.CompareAndSwapPointer(&tw.overflow, nil, unsafe.Pointer(ntw))

			// Load safe to avoid concurrent operations.
			overflow = atomic.LoadPointer(&tw.overflow)
		}

		return (*TimeWheel)(overflow).add(t)
	}
}
