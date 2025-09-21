package v1

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
		// Check if the file type is supported
		if !config.CheckSupportedTypes(ext[1:]) {
			log.Error("Unsupported file type", zap.String("file_type", ext))
			response.BadRequest(w, r, fmt.Errorf("Unsupported file type: %s", ext[1:]))
			return
		}
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
// job -> upload -> parse -> return metadata
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
	if !config.CheckSupportedTypes(ext[1:]) {
		log.Error("Unsupported file type", zap.String("file_type", ext))
		response.BadRequest(w, r, fmt.Errorf("Unsupported file type"))
		return
	}

	bookFileName := strings.TrimSuffix(fileBase, ext)
	log.Debug("Book file name", zap.String("file_name", bookFileName))
	bookPath := fmt.Sprintf("%s/%d/books/%s", config.Opts.Data, uid, bookFileName)
	bookPath = util.GenerateNewDirName(bookPath)
	log.Debug("Book path", zap.String("path", bookPath))
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
	select {
	case bookMeta := <-worker.MetaSingle:
		log.Debug("MetaSingleDone")

		// When We parse the book, we need to save the book metadata
		// Save the book metadata
		newBook := &model.Book{
			ID:           bookMeta.Book.ID,
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
	case err := <-worker.ErrorChan:
		response.ServerError(w, r, err)
		return
	}
}

func (h *Handler) importBooks(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("archive") // Expect a form field named "archive"
	if err != nil {
		response.BadRequest(w, r, errors.New("missing 'archive' file in request"))
		return
	}
	defer file.Close()

	mapTags := r.FormValue("mapTags") == "true"

	uid, _ := strconv.Atoi(request.GetUserID(r))

	// Save the archive to a temporary location
	tmpDir := filepath.Join(config.Opts.Data, fmt.Sprintf("%d/tmp", uid))
	if err := os.MkdirAll(tmpDir, os.ModePerm); err != nil {
		response.ServerError(w, r, errors.Wrap(err, "could not create temp dir for import"))
		return
	}

	archivePath := filepath.Join(tmpDir, header.Filename)
	dst, err := os.Create(archivePath)
	if err != nil {
		response.ServerError(w, r, errors.Wrap(err, "could not save import archive"))
		return
	}

	if _, err := io.Copy(dst, file); err != nil {
		dst.Close()
		response.ServerError(w, r, errors.Wrap(err, "could not write import archive"))
		return
	}
	dst.Close()

	// Launch the processing in a background goroutine.
	// The API returns immediately.
	go h.processImportArchive(archivePath, uid, mapTags)

	log.Info("Book import job accepted", zap.Int("uid", uid), zap.String("archive", header.Filename))
	response.Accepted(w, r) // Respond with 202 Accepted
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

func (h *Handler) getCover(w http.ResponseWriter, r *http.Request) {
	bookID := request.RouteIntParam(r, "bookID")

	book, err := h.store.GetBook(&model.FindBook{BookID: &bookID})
	if err != nil {
		log.Error("Failed to get book", zap.Int("book_id", bookID), zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	dir := filepath.Dir(book.Path)

	log.Debug("Book dir", zap.String("dir", dir))

	var covers []string
	if book.HasCover {
		covers, err = filepath.Glob(fmt.Sprintf("%s/cover.webp", dir))
		if err != nil || len(covers) == 0 {
			log.Error("Failed to get cover", zap.Int("book_id", bookID), zap.String("cover_dir", dir), zap.Error(err))
			http.ServeFile(w, r, fmt.Sprintf("%s/default_cover.webp", "static/img"))
			return
		}
		http.ServeFile(w, r, covers[0])
	} else {
		// Use default cover
		log.Debug("Use default cover", zap.Int("book_id", bookID))
		http.ServeFile(w, r, fmt.Sprintf("%s/default_cover.webp", "static/img"))
	}
}

type addTagRequest struct {
	TagName string `json:"tagName"`
}

func (h *Handler) addTagToBook(w http.ResponseWriter, r *http.Request) {
	// Get the book ID from the URL.
	bookID := request.RouteIntParam(r, "id")
	if bookID == 0 {
		response.BadRequest(w, r, errors.New("invalid book ID"))
		return
	}

	// Decode the incoming JSON request body.
	var req addTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, r, errors.New("invalid request body"))
		return
	}

	// Validate the input.
	tagName := strings.TrimSpace(req.TagName)
	if tagName == "" {
		response.BadRequest(w, r, errors.New("tagName cannot be empty"))
		return
	}

	if err := h.store.AddTagToBook(bookID, tagName); err != nil {
		log.Logger.Error("failed to add tag to book", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	response.OK(w, r, map[string]string{"message": "tag added successfully"})
}

// processImportArchive runs in the background to unpack and import books.
func (h *Handler) processImportArchive(archivePath string, userID int, mapTags bool) {
	// Clean up the archive file when done
	defer os.Remove(archivePath)

	log.Debug("Starting archive processing", zap.String("archive", archivePath))

	archiveFile, err := os.Open(archivePath)
	if err != nil {
		log.Error("Failed to open archive for processing", zap.Error(err))
		return
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		log.Error("Failed to create gzip reader for archive", zap.Error(err))
		return
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Error("Error reading tar header in archive", zap.String("archive", archivePath), zap.Error(err))
			continue
		}

		// Skip directories and unsupported files
		ext := filepath.Ext(header.Name)
		if header.Typeflag != tar.TypeReg || !config.CheckSupportedTypes(ext[1:]) {
			continue
		}

		var tagsToAdd []string
		if mapTags {
			// Get the directory part of the file's path within the archive
			archiveDir := filepath.Dir(header.Name)
			if archiveDir != "." {
				// Split the path by the separator to get individual directory names
				tagsToAdd = strings.Split(archiveDir, string(filepath.Separator))
			}
		}

		// Determine final path for the book
		bookDir := strings.TrimSuffix(filepath.Base(header.Name), ext)
		finalBookDir := fmt.Sprintf("%s/%d/books/%s", config.Opts.Data, userID, bookDir)
		finalBookDir = util.GenerateNewDirName(finalBookDir) // Ensure unique directory

		if err := os.MkdirAll(finalBookDir, os.ModePerm); err != nil {
			log.Error("Failed to create book directory during import", zap.Error(err))
			continue
		}

		finalBookPath := filepath.Join(finalBookDir, filepath.Base(header.Name))

		// Stream the file from the archive to its final destination
		outFile, err := os.Create(finalBookPath)
		if err != nil {
			log.Error("Failed to create book file during import", zap.Error(err))
			continue
		}

		if _, err := io.Copy(outFile, tarReader); err != nil {
			log.Error("Failed to write book file during import", zap.Error(err))
			outFile.Close()
			continue
		}
		outFile.Close()

		// Now that the file is saved, parse its metadata and save it to the DB.
		log.Debug("Imported book saved, now parsing", zap.String("path", finalBookPath))
		if err := h.store.ParseAndSaveBookMeta(finalBookPath, userID, tagsToAdd); err != nil {
			log.Error("Failed to parse and save metadata for imported book", zap.String("path", finalBookPath), zap.Error(err))
			// Optional: Clean up the saved book file on error
			os.RemoveAll(finalBookDir)
		}
	}

	log.Info("Finished processing archive", zap.String("archive", archivePath))
}
