package worker

import (
	"github.com/Xunop/e-oasis/model"
)

type WorkPool interface {
	Push(job model.Job)
}
