package timewheel

import (
	"sync"
	"testing"
	"time"
)

func genD(i int) time.Duration {
	return time.Duration(i%10000) * time.Millisecond
}

type Task3 struct {
	interval time.Duration
	mu       *sync.Mutex
}

func (task *Task3) Next(prev time.Time) time.Time {
	task.mu.Lock()
	defer task.mu.Unlock()
	return prev.Add(task.interval)
}

func (task *Task3) Run() {

}

//func BenchmarkTimeWheel_Schedule(b *testing.B) {
//	tw := New(time.Millisecond, 3)
//	tw.Start()
//	defer tw.Stop()
//
//	for i := 0; i < b.N; i++ {
//		tw.Schedule(&Task3{
//			interval: genD(i),
//			mu:       new(sync.Mutex),
//		})
//	}
//}

func BenchmarkTimeWheel_AfterFunc(b *testing.B) {
	tw := New(time.Millisecond, 3)
	tw.Start()
	defer tw.Stop()

	b.Run("tw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tw.AfterFunc(genD(i), func() {})
		}
	})
	b.Run("std", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			time.AfterFunc(genD(i), func() {})
		}
	})
}
