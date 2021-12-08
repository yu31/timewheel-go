package timewheel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_truncate(t *testing.T) {
	now := time.Now()

	t.Run("1ms", func(t *testing.T) {
		tick := time.Millisecond
		t1 := now.Truncate(tick)
		t2 := time.Unix(0, truncate(now.UnixNano(), int64(tick)))
		require.Equal(t, t1, t2)
	})

	t.Run("5ms", func(t *testing.T) {
		tick := time.Millisecond * 5
		t1 := now.Truncate(tick)
		t2 := time.Unix(0, truncate(now.UnixNano(), int64(tick)))
		require.Equal(t, t1, t2)
	})

	t.Run("10ms", func(t *testing.T) {
		tick := time.Millisecond * 5
		t1 := now.Truncate(tick).UnixNano()
		t2 := truncate(now.UnixNano(), int64(tick))
		require.Equal(t, t1, t2)
	})
}
