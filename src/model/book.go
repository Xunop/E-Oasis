package model

type Meta struct {
}

type Book struct {
	ID           int  `json:"id"`
	Title        string `json:"title"`
	SortTitle    string `json:"sort"`
	TimeStamp    string `json:"timestamp"`
	PublishDate  string `json:"pubdate"`
	SeriesIndex  int  `json:"series_index"`
	AuthorSort   string `json:"author_sort"`
	Isbn         string `json:"isbn"`
	Iccn         string `json:"iccn"`
	Path         string `json:"path"`
	Flags        int    `json:"flags"`
	UUID         string `json:"uuid"`
	HasCover     int    `json:"has_cover"`
	LastModified string `json:"last_modified"`
}

type FindBook struct {
	ID    *int  `json:"id"`
	Title *string `json:"title"`
	// SortTitle string `json:"sort"`
	AuthorSort *string `json:"author_sort"`
	Isbn       *string `json:"isbn"`
	Iccn       *string `json:"iccn"`
	OrderBy    *string `json:"order_by"`

	// Random and limit are used in list books.
	// Whether to return random books.
	Random bool `json:"random"`
	// The maximum number of books to return.
	Limit  *int `json:"limit"`
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
