package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/util"
	"github.com/Xunop/e-oasis/worker"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (h *Handler) listBooks(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(request.GetUserID(r))
	if err != nil {
		log.Logger.Error("Failed to get user ID", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}
	find := &model.FindBook{}
	// If user is not admin or host, only show own books
	if request.GetUserRole(r) != model.RoleHost && request.GetUserRole(r) != model.RoleAdmin {
		log.Debug("User is not admin or host, only show own books")
		find.UserID = &userID
	}

	books, err := h.store.ListBooks(find)
	if err != nil {
		log.Logger.Error("Error listing books", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	response.OK(w, r, books)
}

// addBookBatch need to parse the format of the book and add it to the store
func (h *Handler) addBookBatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(config.Opts.MaxUploadSize << 20); err != nil {
		log.Error("Max upload size exceeded", zap.Error(err))
		log.Error("Now size is", zap.Int64("size", r.ContentLength<<20))
		response.BadRequest(w, r, err)
		return
	}

	files := r.MultipartForm.File["file"]

	jobs := make([]model.Job, 0)
	for _, file := range files {
		uid, err := strconv.Atoi(request.GetUserID(r))
		if err != nil {
			log.Error("Filed to get user ID", zap.Error(err))
			response.BadRequest(w, r, err)
		}

		fileBase := filepath.Base(file.Filename)
		ext := filepath.Ext(fileBase)
		bookDir := strings.TrimSuffix(fileBase, ext)
		bookPath := fmt.Sprintf("%s/%d/books/%s", config.Opts.Data, uid, bookDir)
		bookPath = util.GenerateNewDirName(bookPath)
		job := model.Job{
			UserID: uid,
			Path:   bookPath,
			Type:   "BATCH",
			Status: model.JobStatusPending,
			Item:   file,
		}
		go h.uploadPool.Push(job)
		newJob, err := h.store.AddJob(job)
		if err != nil {
			log.Error("Failed to add job", zap.Error(err))
			response.ServerError(w, r, err)
			return
		}
		jobs = append(jobs, *newJob)
	}
	response.OK(w, r, jobs)
}

// addBookSingle parse the book and return to user
// User can modify book metadata(title, author, cover, etc), so when we parse book, we need to return metadata to user.
// Besides, we can batch upload books, user don't need to modify book metadata.
func (h *Handler) addBookSingle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(config.Opts.MaxUploadSize << 20); err != nil {
		log.Error("Max upload size exceeded", zap.Error(err))
		log.Error("Now size is", zap.Int64("size", r.ContentLength<<20))
		response.BadRequest(w, r, err)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) > 1 || len(files) <= 0 {
		response.BadRequest(w, r, fmt.Errorf("Only one file is allowed"))
		return
	}
	uid, err := strconv.Atoi(request.GetUserID(r))
	if err != nil {
		log.Error("Filed to get user ID", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	// Check if the file type is supported
	fileBase := filepath.Base(files[0].Filename)
	ext := filepath.Ext(fileBase)
	ext = ext[1:]
	if !config.CheckSupportedTypes(ext) {
		log.Error("Unsupported file type", zap.String("file_type", ext))
		response.BadRequest(w, r, fmt.Errorf("Unsupported file type"))
		return
	}

	bookFileName := strings.TrimSuffix(fileBase, ext)
	bookPath := fmt.Sprintf("%s/%d/books/%s", config.Opts.Data, uid, bookFileName)
	bookPath = util.GenerateNewDirName(bookPath)
	job := model.Job{
		UserID: uid,
		Path:   bookPath,
		Type:   "SINGLE",
		Status: model.JobStatusPending,
		Item:   files[0],
	}
	go h.uploadPool.Push(job)
	_, err = h.store.AddJob(job)
	if err != nil {
		log.Error("Failed to add job", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	// FIXME: This is a blocking operation, need to be optimized
	// If goruntine has error, we can't catch it
	// select {
	// case <-worker.MetaSingle:
	// 	log.Debug("MetaSingleDone")
	// case <-worker.MetaSingleError:
	// 	log.Error("MetaSingleError")
	// 	response.ServerError(w, r, errors.New("Failed to parse book"))
	// 	return
	// }
	bookMeta := <-worker.MetaSingle

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
	response.OK(w, r, newBook)
	return
}

// TODO: Add batch delete and delete link data
func (h *Handler) deleteBook(w http.ResponseWriter, r *http.Request) {
	bookID := request.RouteIntParam(r, "id")
	userID, err := strconv.Atoi(request.GetUserID(r))
	if err != nil {
		log.Error("Failed to get user ID", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	log.Debug("Deleting book", zap.Int("bookID", bookID), zap.Int("userID", userID))
	find := &model.FindBook{BookID: &bookID}
	// If user is not admin or host, only allow to delete own book
	if request.GetUserRole(r) != model.RoleHost && request.GetUserRole(r) != model.RoleAdmin {
		find.UserID = &userID
	}

	if err := h.store.RemoveBook(find); err != nil {
		log.Error("Failed to delete book", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	// Delete book from cache
	h.store.BookCache.Delete(bookID)

	response.NoContent(w, r)
}

func (h *Handler) upsetBookStatus(w http.ResponseWriter, r *http.Request) {
	var status model.BookReadingStatusLink
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		log.Error("Failed to decode request body", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}
	// userID is the user who set the status
	userID, err := strconv.Atoi(request.GetUserID(r))
	if err != nil {
		log.Error("Failed to get user ID", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	// Check if the user ID match
	statusUserID := request.RouteIntParam(r, "userID")
	if userID != statusUserID {
		log.Error("User ID not match", zap.Int("userID", userID), zap.Int("requestUserID", statusUserID))
		response.BadRequest(w, r, errors.New("User ID not match"))
		return
	}
	status.UserID = userID

	// Check if the book exists
	bookID := request.RouteIntParam(r, "bookID")
	status.BookID = bookID
	if !h.store.CheckBook(status.BookID) {
		log.Debug("Book not found", zap.Int("bookID", status.BookID))
		response.BadRequest(w, r, errors.New("Book not found"))
		return
	}

	newStatus, err := h.store.UpsetBookStatus(&status)
	if err != nil {
		log.Error("Failed to set book status", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	if newStatus == nil {
		response.BadRequest(w, r, errors.New("Failed to set book status"))
		return
	}

	response.OK(w, r, newStatus)
}

func (h *Handler) getBookStatus(w http.ResponseWriter, r *http.Request) {
    bookID := request.RouteIntParam(r, "bookID")
    userID := request.RouteIntParam(r, "userID")

    currentUserID, err := strconv.Atoi(request.GetUserID(r))
    if err != nil {
		log.Error("Failed to get user ID", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	if currentUserID != userID {
		log.Error("User ID not match", zap.Int("userID", currentUserID), zap.Int("requestUserID", userID))
		response.BadRequest(w, r, errors.New("User ID not match"))
		return
	}

	status, err := h.store.GetBookStatus(bookID, userID)
	if err != nil {
		log.Error("Failed to get book status", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	response.OK(w, r, status)
}
