package worker

import (
	"github.com/Xunop/e-oasis/model"
)

type Worker interface {
	Run(c <-chan model.Job)
}
