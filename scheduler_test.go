package timewheel

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TaskZero1 struct {
}

func (task *TaskZero1) Next(prev time.Time) time.Time {
	return time.Time{}
}

func (task *TaskZero1) Run() {
}

func TestTimeWheel_Schedule_Zero(t *testing.T) {
	tw := Default()
	timer := tw.Schedule(&TaskZero1{})
	require.NotNil(t, timer)
	require.Equal(t, timer.expiration, int64(0))
	require.Nil(t, timer.task)
	require.True(t, timer.b == nil)
	require.Nil(t, timer.element)
	timer.Close()
}
