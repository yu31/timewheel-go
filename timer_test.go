package timewheel

import (
	"context"
	"testing"

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

func TestTimer_Close_With_AfterFunc(t *testing.T) {
	// TODO:
}

func TestTimer_Close_With_Schedule(t *testing.T) {
	// TODO:
}
