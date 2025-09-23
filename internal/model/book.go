package model //import "github.com/Xunop/e-oasis/internal/model"

type Meta struct {
}

type Tag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	BookCount int    `json:"book_count"` // To hold the count of books
}

type Book struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	SortTitle    string `json:"sort"`
	TimeStamp    string `json:"timestamp"`
	PublishDate  string `json:"pubdate"`
	SeriesIndex  int    `json:"series_index"`
	AuthorSort   string `json:"author_sort"`
	ISBN         string `json:"isbn"`
	LCCN         string `json:"lccn"`
	Path         string `json:"path"`
	Flags        int    `json:"flags"`
	UUID         string `json:"uuid"`
	HasCover     bool   `json:"has_cover"`
	LastModified string `json:"last_modified"`
}

type FindBook struct {
	Title  *string `json:"title"`
	UserID *int    `json:"user_id"`
	BookID *int    `json:"book_id"`
	// SortTitle string `json:"sort"`
	AuthorSort *string `json:"author_sort"`
	ISBN       *string `json:"isbn"`
	LCCN       *string `json:"lccn"`
	OrderBy    *string `json:"order_by"`

	// Random and limit are used in list books.
	// Whether to return random books.
	Random bool `json:"random"`
	// The maximum number of books to return.
	Limit *int `json:"limit"`
}

type Publisher struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Sort string `json:"sort"`
}

type Language struct {
	ID       int    `json:"id"`
	LangCode string `json:"lang_code"`
}

type BookMeta struct {
	Book      *Book      `json:"book"`
	Publisher *Publisher `json:"publisher"`
	Language  *Language  `json:"language"`
	Author    *Author    `json:"author"`
}

type BookUserLink struct {
	ID     int `json:"id"`
	BookID int `json:"book"`
	UserID int `json:"user"`
}

type BookAuthorLink struct {
	ID       int `json:"id"`
	BookID   int `json:"book"`
	AuthorID int `json:"author"`
}

type BookLanguagesLink struct {
	ID        int    `json:"id"`
	BookID    int    `json:"book"`
	LangCode  string `json:"lang_code"`
	ItemOrder int    `json:"item_order"`
}

type BookPublishersLink struct {
	ID          int `json:"id"`
	BookID      int `json:"book"`
	PublisherID int `json:"publisher"`
}

type BookTagsLink struct {
	ID     int `json:"id"`
	BookID int `json:"book"`
	TagID  int `json:"tag"`
}

type BookRatingLink struct {
	ID       int `json:"id"`
	BookID   int `json:"book"`
	RatingID int `json:"rating"`
}

type BookSeriesLink struct {
	ID       int `json:"id"`
	BookID   int `json:"book"`
	SeriesID int `json:"series"`
}

type BookPluginLink struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	BookID int    `json:"book"`
	Value  string `json:"val"`
}

// BookReadingStatusLink represents the reading status of a book for a user.
type BookReadingStatusLink struct {
	ID int `json:"id"`
	// The book ID.
	BookID int `json:"book"`
	// The user ID.
	UserID int `json:"user"`
	// The reading status.
	Status int `json:"status"`
	// The percentage of the book read.
	Percentage int `json:"percentage"`
	// The page number of the book.
	Page int `json:"page"`
	// The CurPage is current reading page.
	CurPage int `json:"cur_page"`
	// The total reading duration in seconds.
	Duration int `json:"duration"`
	// The LastRead is the last time the book was read.
    LastRead string `json:"last_read"`
}
