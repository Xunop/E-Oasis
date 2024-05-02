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

// addBook need to parse the format of the book and add it to the store
func (h *Handler) addBook(w http.ResponseWriter, r *http.Request) {
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

// TODO: Add batch delete
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

	response.NoContent(w, r)
}
