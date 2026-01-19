package koduck

import (
	"sync"
	"time"
)

type Scheduler struct {
	tasks []*SchedulerTask
	mu    sync.RWMutex
	quit  chan struct{}
}

type SchedulerTask struct {
	Interval time.Duration
	Next     int64
	Handler  func()
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make([]*SchedulerTask, 0),
		quit:  make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.tick()
			case <-s.quit:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.quit)
}

func (s *Scheduler) tick() {
	now := time.Now().UnixNano()

	s.mu.RLock()
	for _, t := range s.tasks {
		if now >= t.Next {
			// 别阻塞 tick，用 goroutine 执行任务
			go func(task *SchedulerTask) {
				defer func() { recover() }()
				task.Handler()
			}(t)

			t.Next = now + int64(t.Interval)
		}
	}
	s.mu.RUnlock()
}

func (s *Scheduler) Every(interval time.Duration, handler func()) {
	s.mu.Lock()
	s.tasks = append(s.tasks, &SchedulerTask{
		Interval: interval,
		Next:     time.Now().UnixNano() + int64(interval),
		Handler:  handler,
	})
	s.mu.Unlock()
}
