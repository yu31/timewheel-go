package timewheel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimer_Close(t *testing.T) {
	b := newBucket()

	timer := &Timer{}
	b.insert(timer)

	require.Equal(t, b.timers.Front(), timer.element)
	require.Equal(t, b.timers.Front().Value.(*Timer), timer)

	require.Panics(t, func() {
		timer.Close()
	})

	ctxCancel, cancelFunc := context.WithCancel(context.Background())
	timer.ctxCancel = ctxCancel
	timer.cancelFunc = cancelFunc

	require.NotPanics(t, func() {
		timer.Close()
	})

	var done bool
	select {
	case <-ctxCancel.Done():
		done = true
	default:
	}
	require.True(t, done)

	require.Equal(t, b.timers.Len(), 0)
}

func TestTimer_Close_With_TimeFunc1(t *testing.T) {
	tw := Default()
	defer tw.Stop()
	tw.Start()

	ctxCancel, cancelFunc := context.WithCancel(context.Background())
	waitRun := make(chan struct{})

	timer := tw.TimeFunc(
		ctxCancel, time.Now(),
		func(ctx context.Context) error {
			close(waitRun)
			ticker := time.NewTimer(time.Second * 1)
		LOOP:
			for {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					cancelFunc()
					break LOOP
				}
			}
			ticker.Stop()
			return nil
		},
	)

	var isRun bool
	select {
	case <-waitRun:
		isRun = true
	case <-time.After(time.Second):
	}
	require.True(t, isRun)

	timer.Close()

	var isDone bool
	select {
	case <-ctxCancel.Done():
		isDone = true
	case <-time.After(time.Second):
	}

	require.True(t, isDone)
}

func TestTimer_Close_With_TimeFunc2(t *testing.T) {
	tw := Default()
	defer tw.Stop()
	tw.Start()

	ctxCancel, cancelFunc := context.WithCancel(context.Background())
	waitRun := make(chan struct{})

	var jobCtx context.Context

	timer := tw.TimeFunc(
		ctxCancel, time.Now(),
		func(ctx context.Context) error {
			ticker := time.NewTimer(time.Second * 1)
		LOOP:
			for {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					jobCtx = ctx
					close(waitRun)
					break LOOP
				}
			}
			ticker.Stop()
			return nil
		},
	)

	cancelFunc()

	var isRun bool
	select {
	case <-waitRun:
		isRun = true
		require.NotNil(t, jobCtx)
	case <-time.After(time.Second):
	}
	require.True(t, isRun)

	var isDone bool
	select {
	case <-jobCtx.Done():
		isDone = true
	case <-time.After(time.Second):
	}
	require.True(t, isDone)

	require.NotPanics(t, func() {
		timer.Close()
	})
}

func TestTimer_Close_With_Schedule1(t *testing.T) {
	tw := Default()
	defer tw.Stop()
	tw.Start()

	waitRun := make(chan struct{})
	ctxCancel, cancelFunc := context.WithCancel(context.Background())

	timer := tw.ScheduleJob(
		ctxCancel,
		ScheduleFunc(func(t time.Time) time.Time {
			return t.Add(time.Millisecond * 10)
		}),
		JobFunc(func(ctx context.Context) error {
			close(waitRun)
			ticker := time.NewTimer(time.Second * 1)
		LOOP:
			for {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					cancelFunc()
					break LOOP
				}
			}
			ticker.Stop()
			return nil
		}),
	)

	var isRun bool
	select {
	case <-waitRun:
		isRun = true
	case <-time.After(time.Second):
	}
	require.True(t, isRun)

	timer.Close()

	var ok bool
	select {
	case <-ctxCancel.Done():
		ok = true
	case <-time.After(time.Second):
	}

	require.True(t, ok)
}

func TestTimer_Close_With_Schedule2(t *testing.T) {
	tw := Default()
	defer tw.Stop()
	tw.Start()

	ctxCancel, cancelFunc := context.WithCancel(context.Background())
	waitRun := make(chan struct{})
	var jobCtx context.Context

	timer := tw.ScheduleJob(
		ctxCancel,
		ScheduleFunc(func(t time.Time) time.Time {
			return t.Add(time.Millisecond * 10)
		}),
		JobFunc(func(ctx context.Context) error {
			ticker := time.NewTimer(time.Second * 1)
		LOOP:
			for {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					jobCtx = ctx
					close(waitRun)
					break LOOP
				}
			}
			ticker.Stop()
			return nil
		}),
	)

	cancelFunc()

	var isRun bool
	select {
	case <-waitRun:
		isRun = true
		require.NotNil(t, jobCtx)
	case <-time.After(time.Second):
	}
	require.True(t, isRun)

	var isDone bool
	select {
	case <-jobCtx.Done():
		isDone = true
	case <-time.After(time.Second):
	}
	require.True(t, isDone)

	require.NotPanics(t, func() {
		timer.Close()
	})
}
