package timewheel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yu31/dqueue"
)

func TestNew(t *testing.T) {
	size := int64(3)
	tick := time.Second
	tw := New(tick, size)
	require.NotNil(t, tw)
	require.Equal(t, tw.tick, int64(tick))
	require.Equal(t, tw.size, size)
	require.Equal(t, tw.span, int64(tick)*size)
	require.Greater(t, tw.current, int64(0))
	require.Equal(t, len(tw.buckets), int(size))
	require.NotNil(t, tw.queue)
	require.True(t, tw.overflow == nil)
}

func TestDefault(t *testing.T) {
	tw := Default()
	require.NotNil(t, tw)
	require.Equal(t, tw.size, defaultSize)
	require.Equal(t, tw.tick, int64(defaultTick))
}

func TestNew_Panic(t *testing.T) {
	require.Panics(t, func() {
		New(time.Millisecond-1, 1)
	})
	require.Panics(t, func() {
		New(time.Millisecond, 0)
	})
}

func Test_newTimeWheel(t *testing.T) {
	start := time.Now().UnixNano()
	tw := newTimeWheel(int64(time.Second), 3, start, dqueue.Default())
	require.Less(t, tw.current, start)
}

func TestTimeWheel_add(t *testing.T) {
	seeds := []time.Duration{
		1 * time.Millisecond,
		10 * time.Millisecond,
		16 * time.Millisecond,
		46 * time.Millisecond,
	}

	tick := int64(time.Millisecond * 5)
	now := time.Now()

	tw := newTimeWheel(tick, 3, now.UnixNano(), dqueue.Default())

	t1 := &Timer{expiration: now.Add(seeds[0]).UnixNano()}
	require.False(t, tw.add(t1))
	require.Nil(t, t1.getBucket())
	require.Nil(t, t1.element)
	require.Nil(t, (*TimeWheel)(tw.overflow))

	t2 := &Timer{expiration: now.Add(seeds[1]).UnixNano()}
	require.True(t, tw.add(t2))
	require.NotNil(t, t2.getBucket())
	require.NotNil(t, t2.element)
	require.Nil(t, (*TimeWheel)(tw.overflow))

	t3 := &Timer{expiration: now.Add(seeds[2]).UnixNano()}
	require.True(t, tw.add(t3))
	require.NotNil(t, t3.getBucket())
	require.NotNil(t, t3.element)
	require.NotNil(t, (*TimeWheel)(tw.overflow))
	require.Nil(t, (*TimeWheel)((*TimeWheel)(tw.overflow).overflow))

	t4 := &Timer{expiration: now.Add(seeds[3]).UnixNano()}
	require.True(t, tw.add(t4))
	require.NotNil(t, t4.getBucket())
	require.NotNil(t, t4.element)
	require.NotNil(t, (*TimeWheel)(tw.overflow))
	require.NotNil(t, (*TimeWheel)((*TimeWheel)(tw.overflow).overflow))
}
