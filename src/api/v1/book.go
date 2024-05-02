package v1

import (
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
	"go.uber.org/zap"
)

func (h *Handler) listBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.store.ListBooks(&model.FindBook{})
	if err != nil {
		log.Logger.Error("Error listing books", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	response.OK(w, r, books)
}

// addBook need to parse the format of the book and add it to the store
func (h *Handler) addBook(w http.ResponseWriter, r *http.Request) {
	// book := &model.Book{}
	// if err := json.NewDecoder(r.Body).Decode(book); err != nil {
	// 	log.Logger.Error("Error decoding book", zap.Error(err))
	// 	http.Error(w, "Error decoding book", http.StatusBadRequest)
	// 	return
	// }
	// if err := h.store.AddBook(book); err != nil {
	// 	log.Logger.Error("Error adding book", zap.Error(err))
	// 	http.Error(w, "Error adding book", http.StatusInternalServerError)
	// 	return
	// }
	// response.OK(w, r, book)

	if err := r.ParseMultipartForm(config.Opts.MaxUploadSize << 20); err != nil {
		log.Error("Max upload size exceeded", zap.Error(err))
		log.Error("Now size is", zap.Int64("size", r.ContentLength<<20))
		response.BadRequest(w, r, err)
		return
	}

	files := r.MultipartForm.File["file"]

	jobs := make([]model.Job, 0)
	for _, file := range files {
		// if file.Size > config.Opts.MaxUploadSize<<20 {
		// 	err := errors.New("File size exceeded")
		// 	log.Error("Max upload size exceeded", zap.Error(err))
		// 	log.Error("Now size is", zap.Int64("size", r.ContentLength<<20))
		// 	response.BadRequest(w, r, err)
		// 	return
		// }

		uid, err := strconv.Atoi(request.GetUserID(r))
		if err != nil {
			log.Error("Filed to get user ID", zap.Error(err))
			response.BadRequest(w, r, err)
		}

		fileBase := filepath.Base(file.Filename)
		ext := filepath.Ext(fileBase)
		bookFileName := strings.TrimSuffix(fileBase, ext)
		bookPath := fmt.Sprintf("%s/%d/books/%s", config.Opts.Data, uid, bookFileName)
		bookPath = util.GenerateNewDirName(bookPath)
		job := model.Job{
			UserID: uid,
			Path:   bookPath,
			Type:   "UPLOAD",
			Status: model.JobStatusPending,
			Item:   file,
		}
		go h.uploadPool.Push(job)
		jobs = append(jobs, job)
	}
	response.OK(w, r, jobs)
}

// parseBook is a helper function to parse the book format
func (h *Handler) parseBook(name string) (*model.Book, error) {
	return nil, nil
}
