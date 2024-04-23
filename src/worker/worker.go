package worker

import (
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
)

// Worker download books in the background.
type Worker struct {
	// Worker ID
	id    int
	store *store.Store
}

// TODO:
func (w *Worker) Run(c <-chan model.Job) {
}
