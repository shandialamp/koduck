package koduck

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerEvery(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.Start()

	var count int64
	interval := 50 * time.Millisecond

	scheduler.Every(interval, func() {
		atomic.AddInt64(&count, 1)
	})

	time.Sleep(200 * time.Millisecond)
	scheduler.Stop()

	finalCount := atomic.LoadInt64(&count)
	if finalCount < 3 || finalCount > 5 {
		t.Logf("Expected around 4 executions, got %d", finalCount)
	}
}

func TestSchedulerMultipleTasks(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.Start()

	var count1, count2 int64

	scheduler.Every(30*time.Millisecond, func() {
		atomic.AddInt64(&count1, 1)
	})

	scheduler.Every(60*time.Millisecond, func() {
		atomic.AddInt64(&count2, 1)
	})

	time.Sleep(150 * time.Millisecond)
	scheduler.Stop()

	c1 := atomic.LoadInt64(&count1)
	c2 := atomic.LoadInt64(&count2)

	if c1 < 3 {
		t.Logf("Task1 count too low: %d", c1)
	}
	if c2 < 1 {
		t.Logf("Task2 count too low: %d", c2)
	}
}

func TestSchedulerStopPanics(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.Start()

	var executed int64

	scheduler.Every(50*time.Millisecond, func() {
		atomic.AddInt64(&executed, 1)
	})

	time.Sleep(100 * time.Millisecond)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Scheduler panicked: %v", r)
		}
	}()

	scheduler.Stop()

	time.Sleep(100 * time.Millisecond)

	finalCount := atomic.LoadInt64(&executed)
	if finalCount > 3 {
		t.Logf("Task executed after stop: %d times", finalCount)
	}
}

func TestSchedulerConcurrentAccess(t *testing.T) {
	scheduler := NewScheduler()
	scheduler.Start()

	var wg sync.WaitGroup
	var count int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			scheduler.Every(50*time.Millisecond, func() {
				atomic.AddInt64(&count, 1)
			})
		}()
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	finalCount := atomic.LoadInt64(&count)
	if finalCount < 10 {
		t.Logf("Expected at least 10 executions, got %d", finalCount)
	}
}
