package model

type Job struct {
	ID     int
	UserID int
	ItemID interface{}
}

type JobList []Job
