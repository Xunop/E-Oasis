package model //import "github.com/Xunop/e-oasis/model"

const (
	JobStatusPending = "pending"
	JobStatusRunning = "running"
	JobStatusDone    = "done"
)

type Job struct {
	ID     int
	UserID int
	Path   string
	Type   string
	Status string
	Item   interface{}
}

type JobList []Job

func (j JobList) Len() int {
	return len(j)
}
