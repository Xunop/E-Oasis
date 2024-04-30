package model //import "github.com/Xunop/e-oasis/model"

type Job struct {
	ID     int
	UserID int
	ItemID interface{}
}

type JobList []Job
