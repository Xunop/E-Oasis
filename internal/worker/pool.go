package worker

import (
	"github.com/Xunop/e-oasis/internal/model"
)

type WorkPool interface {
	Push(job model.Job)
}
