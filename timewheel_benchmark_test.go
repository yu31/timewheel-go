package timewheel

import (
	"testing"
	"time"
)

func genInterval(i int) time.Duration {
	return time.Duration((i%10000)+1) * time.Millisecond
}

type TaskBench1 struct {
	interval time.Duration
}

func (task *TaskBench1) Next(prev time.Time) time.Time {
	return prev.Add(task.interval)
}

func (task *TaskBench1) Run() error {
	return nil
}

func BenchmarkTimeWheel_Schedule(b *testing.B) {
	tw := New(time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	for i := 0; i < b.N; i++ {
		tw.Schedule(&TaskBench1{
			interval: genInterval(i),
		})
	}
}

func BenchmarkTimeWheel_AfterFunc(b *testing.B) {
	tw := New(time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	b.Run("tw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tw.AfterFunc(genInterval(i), func() error { return nil })
		}
	})
	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.AfterFunc(genInterval(i), func() {})
		}
	})
}

func BenchmarkTimer_StartClose(b *testing.B) {
	tw := New(time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	b.Run("tw", func(b *testing.B) {
		timers := make([]*Timer, 0, b.N)

		for i := 0; i < b.N; i++ {
			timer := tw.AfterFunc(genInterval(i), func() error { return nil })
			timers = append(timers, timer)
		}
		for i := 0; i < b.N; i++ {
			timers[i].Close()
		}
	})
	b.Run("std", func(b *testing.B) {
		timers := make([]*time.Timer, 0, b.N)

		for i := 0; i < b.N; i++ {
			timer := time.AfterFunc(genInterval(i), func() {})
			timers = append(timers, timer)
		}
		for i := 0; i < b.N; i++ {
			timers[i].Stop()
		}
	})
}
