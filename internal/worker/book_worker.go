package worker // import "github.com/Xunop/e-oasis/worker"

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/util/parsers/epub"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	uploadDone = make(chan model.Job)
	metaBatch  = make(chan *model.BookMeta, 10)
	MetaSingle = make(chan model.BookMeta, 1)
	ErrorChan  = make(chan error, 1)
)

type BookUploadPool struct {
	queue chan model.Job
}

func NewUploadPool(store *store.Store, size int) *BookUploadPool {
	pool := &BookUploadPool{
		queue: make(chan model.Job),
	}

	for i := 0; i < size; i++ {
		worker := &BookUploadWorker{id: i, store: store}
		go worker.Run(pool.queue)
	}

	return pool
}

func (p *BookUploadPool) GetQueue() chan model.Job {
	return p.queue
}

// Implement WorkPool interface
func (p *BookUploadPool) Push(job model.Job) {
	p.queue <- job
}

type BookUploadWorker struct {
	id    int
	store *store.Store
}

// Run handler book upload
func (w *BookUploadWorker) Run(c <-chan model.Job) {
	log.Debug("BookDownloadWorker is running", zap.Int("worker_id", w.id))

	for {
		job := <-c

		log.Debug("Job reveived by worker",
			zap.Int("work_id", w.id),
			zap.Int("user_id", job.UserID))

		// startTiime := time.Now()

		uid := job.UserID
		fileHeader := job.Item.(*multipart.FileHeader)
		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			log.Error("Error opening file", zap.Error(err))
			ErrorChan <- err
			continue
		}
		defer file.Close()

		buff := make([]byte, fileHeader.Size)
		_, err = file.Read(buff)
		if err != nil {
			log.Error("Error reading file", zap.Error(err))
			ErrorChan <- err
			continue
		}

		fileType := http.DetectContentType(buff)
		if !config.CheckSupportedTypes(fileType) {
			log.Error("Unsupported file type", zap.String("file_type", fileType))
			ErrorChan <- err
			continue
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			log.Error("Error seeking file", zap.Error(err))
			ErrorChan <- err
			continue
		}

		// Check if the user has a folder
		if _, err := os.Stat(job.Path); os.IsNotExist(err) {
			err = os.MkdirAll(job.Path, os.ModePerm)
			if err != nil {
				log.Error("Error creating folder", zap.Error(err))
				ErrorChan <- err
				continue
			}
		}

		filePath := fmt.Sprintf("%s/%s", job.Path, job.Item.(*multipart.FileHeader).Filename)

		log.Debug("File path", zap.String("path", filePath))

		f, err := os.Create(filePath)
		if err != nil {
			log.Error("Error creating file", zap.Error(err))
			ErrorChan <- err
			os.RemoveAll(job.Path)
			continue
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			log.Error("Error copying file", zap.Error(err))
			ErrorChan <- err
			os.RemoveAll(job.Path)
			continue
		}

		job.Status = model.JobStatusDone
		j, err := w.store.AddJob(job)
		if err != nil {
			log.Error("Error adding job", zap.Error(err))
			// ErrorChan <- err
			// continue
		} else {
			w.store.JobCache.Store(j.ID, &j)
		}

		// Next Parse the book
		// File path is the path of the book: /path/uid/books/book.epub
		uploadDone <- job

		log.Debug("File uploaded successfully",
			zap.String("file_name", fileHeader.Filename),
			zap.Int("user_id", uid),
			zap.Int("job_id", j.ID))
	}
}

type BookParsePool struct {
	Queue chan model.Job
}

// Implement WorkPool interface
func (p *BookParsePool) Push(job model.Job) {
	p.Queue <- job
}

func NewParsePool(store *store.Store, size int) *BookParsePool {
	pool := &BookParsePool{
		Queue: make(chan model.Job),
	}

	go SaveBookMeta(store)
	for i := 0; i < size; i++ {
		worker := &BookParseWorker{id: i, store: store}
		go worker.Run()
	}

	return pool
}

type BookParseWorker struct {
	id    int
	store *store.Store
}

// TODO: Sqlite is in single thread mode, so we can't use multiple thread to insert data
// Can add a pool to handle the insert
// Now use lock to handle the insert
func (w *BookParseWorker) Run() {
	log.Debug("BookParseWorker is running", zap.Int("worker_id", w.id))

	for {
		job := <-uploadDone

		log.Debug("Job reveived by worker",
			zap.Int("work_id", w.id),
			zap.String("path", job.Path))

		log.Debug("Parse book in", zap.String("dir", job.Path))

		filePath := fmt.Sprintf("%s/%s", job.Path, job.Item.(*multipart.FileHeader).Filename)
		bookMeta, err := ParseBook(filePath)
		if err != nil {
			log.Error("Parse book error", zap.String("Book", filePath), zap.Error(err))
			ErrorChan <- err
			os.RemoveAll(job.Path)
			continue
		}

		// When We parse the book, we need to save the book metadata
		// Save the book metadata
		newBook := &model.Book{
			Title:        bookMeta.Book.Title,
			SortTitle:    bookMeta.Book.SortTitle,
			PublishDate:  bookMeta.Book.PublishDate,
			AuthorSort:   bookMeta.Book.AuthorSort,
			ISBN:         bookMeta.Book.ISBN,
			Path:         bookMeta.Book.Path,
			UUID:         bookMeta.Book.UUID,
			HasCover:     bookMeta.Book.HasCover,
			LastModified: bookMeta.Book.LastModified,
		}

		returnBook, err := w.store.AddBook(newBook)
		if err != nil {
			log.Error("Error adding book", zap.Error(err))
			ErrorChan <- err
			continue
		}
		w.store.BookCache.Store(returnBook.ID, returnBook)
		bookMeta.Book = returnBook
		if job.Type == "SINGLE" {
			MetaSingle <- *bookMeta
		}

		metaBatch <- bookMeta
	}
}

func SaveBookMeta(s *store.Store) {
	for {
		metaData := <-metaBatch

		// When We parse the book, we need to save the book metadata
		// Save the book metadata
		// newBook := &model.Book{
		// 	Title:        metaData.Book.Title,
		// 	SortTitle:    metaData.Book.SortTitle,
		// 	PublishDate:  metaData.Book.PublishDate,
		// 	AuthorSort:   metaData.Book.AuthorSort,
		// 	ISBN:         metaData.Book.ISBN,
		// 	Path:         metaData.Book.Path,
		// 	UUID:         metaData.Book.UUID,
		// 	HasCover:     metaData.Book.HasCover,
		// 	LastModified: metaData.Book.LastModified,
		// }
		//
		// returnBook, err := s.AddBook(newBook)
		// if err != nil {
		// 	log.Error("Error adding book", zap.Error(err))
		// 	ErrorChan <- err
		// 	continue
		// }
		// s.BookCache.Store(returnBook.ID, returnBook)

		returnBook := metaData.Book

		publisherRes, err := s.AddPublisher(metaData.Publisher)
		if err != nil {
			log.Error("Error add publisher", zap.String("publisher", metaData.Publisher.Name), zap.Error(err))
		}
		log.Debug("Add publisher response", zap.Any("response", publisherRes))

		authorRes, err := s.AddAuthor(metaData.Author)
		if err != nil {
			log.Error("Error add author", zap.String("author", metaData.Author.Name), zap.Error(err))
		}
		log.Debug("Add author response", zap.Any("response", authorRes))

		authorLink := model.BookAuthorLink{BookID: returnBook.ID, AuthorID: authorRes.ID}
		linkRes, err := s.AddBookAuthorLink(&authorLink)
		if err != nil {
			log.Error("Error add book author link", zap.Error(err))
		}
		log.Debug("Add book author link response", zap.Any("response", linkRes))

		uidIdx := 0
		for idx, part := range strings.Split(metaData.Book.Path, "/") {
			if part == "books" {
				uidIdx = idx
			}
		}
		if uidIdx == 0 {
			log.Error("Error getting user ID", zap.String("path", metaData.Book.Path))
			continue
		}
		uid, err := strconv.Atoi(strings.Split(metaData.Book.Path, "/")[uidIdx-1])
		if err != nil {
			log.Error("Error getting user ID", zap.Error(err), zap.String("path", metaData.Book.Path))
		}
		bookUserLinkRes, err := s.AddBookUserLink(&model.BookUserLink{BookID: returnBook.ID, UserID: uid})
		if err != nil {
			log.Error("Error add book user link", zap.Error(err))
		}
		log.Debug("Add book user link response", zap.Any("response", bookUserLinkRes))
		// _, err := w.store.AddLanguage(&model.Language{Name: bookLanguage})
		// w.store.AddBookAuthorLink(&model.BookAuthorLink{BookID: returnBook.ID, AuthorID: 1})
	}
}

// generateBookHash generate the hash of the book
func generateBookHash(book string) (string, error) {
	bookType := filepath.Ext(book)
	switch bookType {
	case ".epub":
		epub, err := epub.Open(book)
		if err != nil {
			log.Error("Error opening epub", zap.Error(err))
			return "", err
		}
		defer epub.Close()

		content := ""

		// Metadata
		content += epub.GetTitle()
		content += epub.GetAuthor()
		content += epub.GetDate()
		content += epub.GetPublisher()
		content += epub.GetDescription()

		items := epub.Opf.Manifest
		for _, item := range items {
			if item.MediaType != "application/xhtml+xml" {
				continue
			}
			data, err := epub.GetContent(item.Href)
			if err != nil {
				return "", err
			}
			content += string(data)
		}

		hash := sha256.New()
		_, err = hash.Write([]byte(content))
		if err != nil {
			log.Error("Error writing hash", zap.Error(err))
			return "", err
		}
		// hash := sha256.Sum256([]byte(content))
		return hex.EncodeToString(hash.Sum(nil)), nil
	default:
		return "", errors.New("Unsupported book type")
	}
}
