package worker

import (
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
)

type Pool struct {
	queue chan model.Job
}

func (p *Pool) Push(jobs model.JobList) {
	for _, job := range jobs {
		p.queue <- job
	}
}

// NewPool creates a pool of background workers.
func NewPool(store *store.Store, size int) *Pool {
	workerPool := &Pool{
		queue: make(chan model.Job),
	}

	for i := 0; i < size; i++ {
		worker := &Worker{id: i, store: store}
		go worker.Run(workerPool.queue)
	}
	return workerPool
}
