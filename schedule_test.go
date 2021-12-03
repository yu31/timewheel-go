package timewheel

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeWheel_expireFunc(t *testing.T) {
	tw := New(time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	seeds := []time.Duration{
		time.Millisecond * 1,
		time.Millisecond * 5,
		time.Millisecond * 10,
		time.Millisecond * 50,
		time.Millisecond * 100,
		time.Millisecond * 400,
		time.Millisecond * 500,
		time.Second * 1,
	}

	for _, d := range seeds {
		t.Run(d.String(), func(t *testing.T) {
			retC := make(chan time.Time)

			start := time.Now()

			min := start
			max := start.Add(d + time.Millisecond*5)

			timer := tw.expireFunc(time.Now().Add(d).UnixNano(), func() error { retC <- time.Now(); return nil })
			require.NotNil(t, timer)

			got := <-retC

			require.Greater(t, got.UnixNano(), min.UnixNano(), fmt.Sprintf("%s: got: %s, want: %s", d.String(), got.String(), min.String()))
			require.Less(t, got.UnixNano(), max.UnixNano(), fmt.Sprintf("%s: got: %s, want: %s", d.String(), got.String(), max.String()))
		})
	}
}

// ScheduleNext for test case TestTimeWheel_Schedule_Next.
type ScheduleNext struct {
	mu    *sync.Mutex
	seeds []time.Duration
	index int
	retC  chan time.Time
}

func (s *ScheduleNext) Next(prev time.Time) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.index >= len(s.seeds) {
		return time.Time{}
	}
	next := prev.Add(s.seeds[s.index])
	s.index += 1
	return next
}

func (s *ScheduleNext) Run() error {
	s.retC <- time.Now()
	return nil
}

func TestTimeWheel_Schedule_Next(t *testing.T) {
	tw := New(time.Millisecond, 20)
	tw.Start()
	defer tw.Stop()

	seeds := []time.Duration{
		1 * time.Millisecond,   // start + 1ms
		4 * time.Millisecond,   // start + 5ms
		5 * time.Millisecond,   // start + 10ms
		40 * time.Millisecond,  // start + 50ms
		50 * time.Millisecond,  // start + 100ms
		400 * time.Millisecond, // start + 400ms
		500 * time.Millisecond, // start + 500ms
		501 * time.Millisecond, // start + 501ms
	}

	retC := make(chan time.Time)

	sh := &ScheduleNext{
		mu:    new(sync.Mutex),
		seeds: seeds,
		retC:  retC,
	}

	lapse := time.Duration(0)
	start := time.Now()

	timer := tw.ScheduleJob(sh, sh)
	require.NotNil(t, timer)

	for _, d := range seeds {
		lapse += d
		min := start
		max := start.Add(lapse + time.Millisecond*5)

		got := <-retC

		require.Greater(t, got.UnixNano(), min.UnixNano(), fmt.Sprintf("%s: got: %s, want: %s", d.String(), got.String(), min.String()))
		require.Less(t, got.UnixNano(), max.UnixNano(), fmt.Sprintf("%s: got: %s, want: %s", d.String(), got.String(), max.String()))
	}
}

// ScheduleRun for test cases TestTimeWheel_Schedule_Run.
type ScheduleRun struct {
	interval time.Duration
	count    int
	mu       *sync.Mutex
	wg       *sync.WaitGroup
	t        *testing.T
	zero     bool
}

func (task *ScheduleRun) Next(prev time.Time) time.Time {
	task.mu.Lock()
	defer task.mu.Unlock()

	require.False(task.t, prev.IsZero())

	if task.count <= 1 {
		task.zero = true
		return time.Time{}
	}
	return prev.Add(task.interval)
}

func (task *ScheduleRun) Run() error {
	task.mu.Lock()
	defer task.mu.Unlock()

	task.count--
	task.wg.Done()
	return nil
}

// For test the previous is not zero in Next.
func TestTimeWheel_Schedule_Run(t *testing.T) {
	task := &ScheduleRun{
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

	timer := tw.ScheduleJob(task, task)
	require.Equal(t, tw.queue.Len(), 1)

	task.wg.Wait()

	require.True(t, task.zero)
	require.Equal(t, task.count, 0)
	// The task not be re-insert to the queue if return zero time in task.Next.
	require.Equal(t, tw.queue.Len(), 0)

	timer.Close()
}

// ScheduleZero for test cases TestTimeWheel_Schedule_Zero
type ScheduleZero struct {
}

func (task *ScheduleZero) Next(prev time.Time) time.Time {
	return time.Time{}
}

func (task *ScheduleZero) Run() error {
	return nil
}

func TestTimeWheel_Schedule_Zero(t *testing.T) {
	task := &ScheduleZero{}
	tw := Default()
	timer := tw.ScheduleJob(task, task)
	require.NotNil(t, timer)
	require.Equal(t, timer.expiration, int64(0))
	require.Nil(t, timer.task)
	require.True(t, timer.b == nil)
	require.Nil(t, timer.element)
	timer.Close()
}
