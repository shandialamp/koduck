package koduck

import "sync"

type WorkerPool struct {
	queue chan func()
	wg    sync.WaitGroup
}

func NewWorkerPool(size int) *WorkerPool {
	pool := &WorkerPool{
		queue: make(chan func(), 1000),
	}
	for i := 0; i < size; i++ {
		go pool.worker()
	}
	return pool
}

func (p *WorkerPool) worker() {
	for task := range p.queue {
		task()
		p.wg.Done()
	}
}

func (p *WorkerPool) Submit(task func()) {
	p.wg.Add(1)
	select {
	case p.queue <- task:
	default:
		// 如果队列满，可以选择丢弃或阻塞
		go task() // 回退策略：直接启动新 goroutine
		p.wg.Done()
	}
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

func (p *WorkerPool) Stop() {
	close(p.queue)
	p.Wait()
}
