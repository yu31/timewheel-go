package timewheel

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yu31/dqueue-go"
)

func TestNew(t *testing.T) {
	size := int64(3)
	tick := time.Second
	tw := New(tick, size)
	require.NotNil(t, tw)
	require.Equal(t, tw.tick, durationToMs(tick))
	require.Equal(t, tw.size, size)
	require.Equal(t, tw.span, durationToMs(tick)*size)
	require.Greater(t, tw.current, int64(0))
	require.Equal(t, len(tw.buckets), int(size))
	require.NotNil(t, tw.queue)
	require.Equal(t, tw.location, time.Local)
	require.Nil(t, tw.opts)
	require.True(t, tw.overflow == nil)
}

func TestDefault(t *testing.T) {
	tw := Default()
	require.NotNil(t, tw)
	require.Equal(t, tw.size, defaultSize)
	require.Equal(t, tw.tick, durationToMs(defaultTick))
	require.Equal(t, tw.location, time.Local)
	require.Nil(t, tw.opts)
}

func TestWithTimezone(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		tw := Default(WithTimezone(time.UTC))
		require.Equal(t, 1, len(tw.opts))
		require.Equal(t, tw.location, time.UTC)
	})
	t.Run("New", func(t *testing.T) {
		tw := New(time.Second, 3, WithTimezone(time.UTC))
		require.Equal(t, 1, len(tw.opts))
		require.Equal(t, tw.location, time.UTC)
	})
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

	t.Run("Default", func(t *testing.T) {
		now := time.Now()
		tw := newTimeWheel(tick, 3, now.UnixNano(), dqueue.Default())

		t1 := &Timer{expiration: now.Add(seeds[0]).UnixNano()}
		require.False(t, tw.add(t1), fmt.Sprintf(
			"now: %s, unixnano: %d, expiration: %d, current: %d, tick: %d",
			now, now.UnixNano(), t1.expiration, tw.current, tw.tick,
		))
		require.Nil(t, t1.getBucket())
		require.Nil(t, t1.element)
		require.Equal(t, tw.location, time.Local)
		p1 := (*TimeWheel)(tw.overflow)
		require.Nil(t, p1)

		t2 := &Timer{expiration: now.Add(seeds[1]).UnixNano()}
		require.True(t, tw.add(t2))
		require.NotNil(t, t2.getBucket())
		require.NotNil(t, t2.element)
		require.Equal(t, tw.location, time.Local)
		p2 := (*TimeWheel)(tw.overflow)
		require.Nil(t, p2)

		t3 := &Timer{expiration: now.Add(seeds[2]).UnixNano()}
		require.True(t, tw.add(t3))
		require.NotNil(t, t3.getBucket())
		require.NotNil(t, t3.element)
		p3 := (*TimeWheel)(tw.overflow)
		require.NotNil(t, p3)
		require.Equal(t, p3.location, time.Local)
		pp3 := (*TimeWheel)(p3.overflow)
		require.Nil(t, pp3)

		t4 := &Timer{expiration: now.Add(seeds[3]).UnixNano()}
		require.True(t, tw.add(t4))
		require.NotNil(t, t4.getBucket())
		require.NotNil(t, t4.element)
		p4 := (*TimeWheel)(tw.overflow)
		require.NotNil(t, p4)
		require.Equal(t, p4.location, time.Local)
		pp4 := (*TimeWheel)(p4.overflow)
		require.NotNil(t, pp4)
		require.Equal(t, pp4.location, time.Local)
		ppp4 := (*TimeWheel)(pp4.overflow)
		require.Nil(t, ppp4)
	})

	t.Run("WithTimezone", func(t *testing.T) {
		now := time.Now()
		tw := newTimeWheel(tick, 3, now.UnixNano(), dqueue.Default(), WithTimezone(time.UTC))

		t1 := &Timer{expiration: now.Add(seeds[0]).UnixNano()}
		require.False(t, tw.add(t1), fmt.Sprintf(
			"now: %s, unixnano: %d, expiration: %d, current: %d, tick: %d",
			now, now.UnixNano(), t1.expiration, tw.current, tw.tick,
		))
		require.Nil(t, t1.getBucket())
		require.Nil(t, t1.element)
		require.Equal(t, tw.location, time.UTC)
		p1 := (*TimeWheel)(tw.overflow)
		require.Nil(t, p1)

		t2 := &Timer{expiration: now.Add(seeds[1]).UnixNano()}
		require.True(t, tw.add(t2))
		require.NotNil(t, t2.getBucket())
		require.NotNil(t, t2.element)
		require.Equal(t, tw.location, time.UTC)
		p2 := (*TimeWheel)(tw.overflow)
		require.Nil(t, p2)

		t3 := &Timer{expiration: now.Add(seeds[2]).UnixNano()}
		require.True(t, tw.add(t3))
		require.NotNil(t, t3.getBucket())
		require.NotNil(t, t3.element)
		p3 := (*TimeWheel)(tw.overflow)
		require.NotNil(t, p3)
		require.Equal(t, p3.location, time.UTC)
		pp3 := (*TimeWheel)(p3.overflow)
		require.Nil(t, pp3)

		t4 := &Timer{expiration: now.Add(seeds[3]).UnixNano()}
		require.True(t, tw.add(t4))
		require.NotNil(t, t4.getBucket())
		require.NotNil(t, t4.element)
		p4 := (*TimeWheel)(tw.overflow)
		require.NotNil(t, p4)
		require.Equal(t, p4.location, time.UTC)
		pp4 := (*TimeWheel)(p4.overflow)
		require.NotNil(t, pp4)
		require.Equal(t, pp4.location, time.UTC)
		ppp4 := (*TimeWheel)(pp4.overflow)
		require.Nil(t, ppp4)
	})
}
