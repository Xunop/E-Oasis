package worker // import "github.com/Xunop/e-oasis/worker"

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
	"go.uber.org/zap"
)

var (
	jobDone = make(chan string)
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
		file, err := fileHeader.Open()
		if err != nil {
			log.Error("Error opening file", zap.Error(err))
			continue
		}
		defer file.Close()

		buff := make([]byte, fileHeader.Size)
		_, err = file.Read(buff)
		if err != nil {
			log.Error("Error reading file", zap.Error(err))
			continue
		}

		fileType := http.DetectContentType(buff)
		supportType := config.Opts.SupportedTypes
		if !isSupportedType(fileType, supportType) {
			log.Error("Unsupported file type", zap.String("file_type", fileType))
			continue
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			log.Error("Error seeking file", zap.Error(err))
			continue
		}

		// Check if the user has a folder
		if _, err := os.Stat(job.Path); os.IsNotExist(err) {
			err = os.MkdirAll(job.Path, os.ModePerm)
		}

		filePath := fmt.Sprintf("%s/%s", job.Path, job.Item.(*multipart.FileHeader).Filename)
		f, err := os.Create(filePath)
		if err != nil {
			log.Error("Error creating file", zap.Error(err))
			continue
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			log.Error("Error copying file", zap.Error(err))
			continue
		}

		job.Status = model.JobStatusDone
		j, err := w.store.AddJob(job)
		if err != nil {
			log.Error("Error adding job", zap.Error(err))
			continue
		}

		w.store.JobCache.Store(j.ID, &j)
		// Next Parse the book
		// File path is the path of the book: /path/to/book/book.epub
		jobDone <- filePath

		log.Debug("File uploaded successfully",
			zap.String("file_name", fileHeader.Filename),
			zap.Int("user_id", uid),
			zap.Int("job_id", j.ID))
	}
}

func isSupportedType(fileType string, supportType []string) bool {
	for _, t := range supportType {
		log.Debug("Checking file type", zap.String("file_type", t), zap.String("supported_type", fileType))
		if t == fileType {
			return true
		}
	}
	return false
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
		path := <-jobDone

		log.Debug("Job reveived by worker",
			zap.Int("work_id", w.id),
			zap.String("path", path))

		bookMeta, err := parseBook(path)
		if err != nil {
			log.Error("Parse book error", zap.String("Book", path), zap.Error(err))
			continue
		}

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
			continue
		}
		w.store.BookCache.Store(returnBook.ID, returnBook)

		// w.store.AddBookAuthorLink(&model.BookAuthorLink{BookID: returnBook.ID, AuthorID: 1})

		// TODO: Handler return publisher
		_, err = w.store.AddPublisher(bookMeta.Publisher)
		if err != nil {
			log.Error("Error add publisher", zap.String("publisher", bookMeta.Publisher.Name), zap.Error(err))
			continue
		}

		// _, err := w.store.AddLanguage(&model.Language{Name: bookLanguage})
	}
}
