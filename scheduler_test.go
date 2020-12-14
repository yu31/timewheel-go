package timewheel

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Task3 struct {
}

func (task *Task3) Next(prev time.Time) time.Time {
	return time.Time{}
}

func (task *Task3) Run() {
}

func TestTimeWheel_Schedule_Zero(t *testing.T) {
	tw := Default()
	timer := tw.Schedule(&Task3{})
	require.NotNil(t, timer)
	require.Equal(t, timer.expiration, int64(0))
	require.Nil(t, timer.task)
	require.True(t, timer.b == nil)
	require.Nil(t, timer.element)
	timer.Close()
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
