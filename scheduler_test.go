package timewheel

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Task3 struct {
	First  time.Duration
	Second time.Duration
	mu     sync.Mutex
	n      int
}

func (task *Task3) Next(prev time.Time) time.Time {
	task.mu.Lock()
	defer task.mu.Unlock()

	if task.n == 0 {
		task.n++
		return prev.Add(task.First)
	}
	task.n++
	return prev.Add(task.Second)
}

func (task *Task3) Run() {
}

func TestTimeWheel_Schedule_Panic(t *testing.T) {
	t.Run("case1", func(t *testing.T) {
		tw := Default()
		tw.Start()
		defer tw.Stop()

		require.Panics(t, func() {
			tw.Schedule(&Task3{First: 0})
		})
		require.Panics(t, func() {
			tw.Schedule(&Task3{First: time.Nanosecond * 10})
		})
		require.Panics(t, func() {
			tw.Schedule(&Task3{First: time.Nanosecond * 999})
		})
		require.NotPanics(t, func() {
			tw.Schedule(&Task3{First: time.Microsecond, Second: time.Microsecond})
		})
	})

	// Should panic.
	t.Run("case2", func(t *testing.T) {
		task := &Task3{
			First:  time.Microsecond,
			Second: time.Nanosecond * 10,
		}
		require.Panics(t, func() {
			tw := Default()
			tw.Start()
			defer tw.Stop()

			tw.Schedule(task)

			tw.Wait()
		})
		require.Equal(t, task.n, 2)
	})

	t.Run("case3", func(t *testing.T) {
		task := &Task3{
			First:  time.Microsecond,
			Second: time.Millisecond,
		}
		require.NotPanics(t, func() {
			tw := Default()
			tw.Start()
			defer tw.Stop()

			tw.Schedule(task)

			time.Sleep(time.Millisecond * 100)
		})
		require.Greater(t, task.n, 1)
	})
}

type Task2 struct {
	interval time.Duration
	count    int
	mu       *sync.Mutex
	wg       *sync.WaitGroup
	t        *testing.T
	zero     bool
}

func (task *Task2) Next(prev time.Time) time.Time {
	task.mu.Lock()
	defer task.mu.Unlock()

	require.False(task.t, prev.IsZero())

	if task.count <= 1 {
		task.zero = true
		return time.Time{}
	}
	return prev.Add(task.interval)
}

func (task *Task2) Run() {
	task.mu.Lock()
	defer task.mu.Unlock()

	task.count--
	task.wg.Done()
}

// For test the previous is not zero in Next.
func TestTimeWheel_Schedule_Next(t *testing.T) {
	task := &Task2{
		interval: time.Millisecond * 5,
		count:    10,
		mu:       new(sync.Mutex),
		wg:       new(sync.WaitGroup),
		t:        t,
		zero:     false,
	}

	task.wg.Add(task.count)

	tw := Default()
	defer tw.Stop()
	tw.Start()

	timer := tw.Schedule(task)
	require.Equal(t, tw.queue.Len(), 1)

	task.wg.Wait()

	require.True(t, task.zero)
	require.Equal(t, task.count, 0)
	// The task not be re-insert to the queue if return zero time in task.Next.
	require.Equal(t, tw.queue.Len(), 0)

	timer.Close()
}
