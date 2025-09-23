package worker

import (
	"github.com/Xunop/e-oasis/internal/model"
)

type Worker interface {
	Run(c <-chan model.Job)
}
