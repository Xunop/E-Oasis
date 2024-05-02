package worker // import "github.com/Xunop/e-oasis/worker"

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/util"
	"github.com/Xunop/e-oasis/util/parsers/epub"
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
		if _, err := os.Stat(filepath.Dir(job.Path)); os.IsNotExist(err) {
			err = os.MkdirAll(filepath.Dir(job.Path), os.ModePerm)
		}

		f, err := os.Create(job.Path)
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

		// Next Parse the book
		jobDone <- job.Path

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
func (w *BookParseWorker) Run() {
	log.Debug("BookParseWorker is running", zap.Int("worker_id", w.id))

	for {
		path := <-jobDone

		log.Debug("Job reveived by worker",
			zap.Int("work_id", w.id),
			zap.String("path", path))

		book, err := epub.Open(path)
		if err != nil {
			log.Error("Error opening epub", zap.Error(err))
			continue
		}

		// Get the book metadata(title, author, etc)
		bookTitle := book.GetTitle()
		bookAuthor := book.GetAuthor()
		bookCover := book.GetCover()
		hasCover := false
		if bookCover != "" {
			hasCover = true
		}
		bookUUID := book.GetUUID()
		bookISBN := book.GetISBN()
		bookDate := book.GetDate()
		bookPublisher := book.GetPublisher()
		// bookLanguage := book.GetLanguage()

		log.Debug("Book parse worker:", zap.String("Book title", bookTitle), zap.String("Book author", bookAuthor))
		sortTitle := util.TitleSort(bookTitle)
		sortAuthor := util.GetSortedAuthor(bookAuthor)
		// log.Debug("Book title: %s, Book author: %s", zap.String("title", sortTitle), zap.String("author", sortAuthor))

		// Save the book metadata
		newBook := &model.Book{
			Title: bookTitle,
			SortTitle:    sortTitle,
			PublishDate: bookDate,
			AuthorSort:   sortAuthor,
			ISBN:         bookISBN,
			Path:         path,
			UUID:         bookUUID,
			HasCover:     hasCover,
			LastModified: time.Now().String(),
		}

		returnBook, err := w.store.AddBook(newBook)
		if err != nil {
			log.Error("Error adding book", zap.Error(err))
			continue
		}
		w.store.BookCache.Store(returnBook.ID, returnBook)

		// w.store.AddBookAuthorLink(&model.BookAuthorLink{BookID: returnBook.ID, AuthorID: 1})

		// TODO: Handler return publisher
		_, err = w.store.AddPublisher(&model.Publisher{Name: bookPublisher})
		if err != nil {
			log.Error("Error add publisher", zap.Error(err))
			continue
		}

		// _, err := w.store.AddLanguage(&model.Language{Name: bookLanguage})
	}
}
