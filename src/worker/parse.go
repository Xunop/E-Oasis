package worker // import "github.com/Xunop/e-oasis/worker"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/util"
	"github.com/Xunop/e-oasis/util/parsers/epub"
	"go.uber.org/zap"
)

func ParseBook(path string) (*model.BookMeta, error) {
	bookType := filepath.Ext(path)
	switch bookType {
	case ".epub":
		return parseEpub(path)
	default:
		return nil, fmt.Errorf("Unsupported file type: %s", bookType)
	}
}

func parseEpub(path string) (*model.BookMeta, error) {
	book, err := epub.Open(path)
	if err != nil {
		log.Error("Error opening epub", zap.Error(err))
		return nil, err
	}
	defer book.Close()

	// Get the book metadata(title, author, etc)
	bookTitle := book.GetTitle()
	bookAuthor := book.GetAuthor()
	hasCover := false
	// Book cover always in book directory, but don't know the extension of the cover(jpg/png?)
	bookCover, err := book.GetCover(filepath.Dir(path))
	if bookCover != "" && err == nil {
		hasCover = true
	}

	// Cover to webp
	// Is reckless to use goroutine here?
	go util.ImageToWebp(bookCover, 75)

	// Remove the original cover
	go os.Remove(bookCover)
	bookUUID := book.GetUUID()
	bookISBN := book.GetISBN()
	bookDate := book.GetDate()
	bookPublisher := book.GetPublisher()
	bookPublisher = strings.TrimSpace(bookPublisher)
	if bookPublisher == "" {
		bookPublisher = "Unknown"
	}
	// bookLanguage := book.GetLanguage()

	log.Debug("Book parse worker:", zap.String("Book title", bookTitle), zap.String("Book author", bookAuthor))

	sortAuthor := util.AuthorSort(bookAuthor)
	sortTitle := util.TitleSort(bookTitle)

	log.Debug("Book title: %s, Book author: %s", zap.String("title", sortTitle), zap.String("author", sortAuthor))

	// Save the book metadata
	newBook := &model.Book{
		Title:        bookTitle,
		SortTitle:    sortTitle,
		PublishDate:  bookDate,
		AuthorSort:   sortAuthor,
		ISBN:         bookISBN,
		Path:         path,
		UUID:         bookUUID,
		HasCover:     hasCover,
		LastModified: time.Now().String(),
	}
	bookMeta := &model.BookMeta{
		Book:      newBook,
		Publisher: &model.Publisher{Name: bookPublisher},
		Language:  &model.Language{},
		Author:    &model.Author{Name: bookAuthor, Sort: sortAuthor},
	}
	return bookMeta, err
}
