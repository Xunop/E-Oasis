package model

// Duration represents every time a user reads a book.
type Duration struct {
	ID           int `json:"id"`
	BookID       int `json:"book_id"`
	UserID       int `json:"user_id"`
	ReadDuration int `json:"read_duration"`
	Percentage   int `json:"percentage"`
	StartTime    int `json:"start_time"`
}
