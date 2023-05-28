// Copyright (c) 2020, Yu Wu <yu.771991@gmail.com> All rights reserved.
//
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package timewheel

import (
	"container/list"
	"context"
	"sync/atomic"
	"unsafe"
)

// Timer represents a single event. The given jobFunc will be executed when the timer expires.
type Timer struct {
	// ctxCancel is created with context.WithCancel
	ctxCancel context.Context

	// cancelFunc is used in Close.
	cancelFunc context.CancelFunc

	// expiration is the expiry time in milliseconds
	expiration int64

	jobFunc JobFunc

	// The bucket that holds the list to which this timer's element belongs.
	//
	// NOTICE: This field may be updated and read concurrently,
	// through Timer.Close() and Bucket.flush().
	b unsafe.Pointer // type: *bucket

	// The timer's Element in list.
	element *list.Element
}

func (t *Timer) getBucket() *bucket {
	return (*bucket)(atomic.LoadPointer(&t.b))
}

func (t *Timer) setBucket(b *bucket) {
	atomic.StorePointer(&t.b, unsafe.Pointer(b))
}

// Close prevents the Timer from firing.
//
// The func will be blocked until the timer has finally been removed from the TimeWheel.
// But, if the timer t has already expired and the t.jobFunc has been started in its own
// goroutine; Close does not wait for t.jobFunc to complete before returning. If the invoker
// needs to know whether t.jobFunc is completed, it must coordinate with t.jobFunc explicitly.
func (t *Timer) Close() {
	for b := t.getBucket(); b != nil; b = t.getBucket() {
		// The b.delete may fail if t's bucket has changed due to TimeWheel call the b.flush.
		// Thus, we re-get t's possibly new bucket and retry until the bucket becomes nil or
		// delete successful.
		if ok := b.delete(t); ok {
			break
		}
	}
	t.cancelFunc()
}
