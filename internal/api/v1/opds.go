package v1

import (
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"go.uber.org/zap"
)

// OpdsTemplateData holds data for the OPDS XML template.
type OpdsTemplateData struct {
	Books       []*OpdsBook
	BaseURL     string
	CurrentTime string
}

// OpdsBook is a simplified book structure for the template.
type OpdsBook struct {
	ID           int
	Title        string
	AuthorSort   string
	LastModified time.Time
	Path         string
	MimeType     string
	Size         int64
	HasCover     bool
}

// opdsFeed is the handler for the main OPDS catalog feed.
func (h *Handler) opdsFeed(w http.ResponseWriter, r *http.Request) {
	// Fetch all books from the store
	books, err := h.store.ListBooks(&model.FindBook{})
	if err != nil {
		log.Logger.Error("failed to list books for OPDS feed", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	// Prepare the data for the template
	opdsBooks := make([]*OpdsBook, 0, len(books))
	for _, book := range books {
		// Get file info for size
		fileInfo, err := os.Stat(book.Path)
		var fileSize int64 = 0
		if err == nil {
			fileSize = fileInfo.Size()
		}

		// Get MIME type from extension
		mimeType := mime.TypeByExtension(filepath.Ext(book.Path))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		
		lastModifiedTime, _ := time.Parse(time.RFC3339, book.LastModified)

		opdsBooks = append(opdsBooks, &OpdsBook{
			ID:           book.ID,
			Title:        book.Title,
			AuthorSort:   book.AuthorSort,
			LastModified: lastModifiedTime,
			Path:         book.Path,
			MimeType:     mimeType,
			Size:         fileSize,
			HasCover:     book.HasCover,
		})
	}

	data := OpdsTemplateData{
		Books:       opdsBooks,
		BaseURL:     getBaseURL(r), // We'll create this helper function
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
	}

	// Parse and execute the XML template
	tmpl, err := template.New("opds.xml").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
	}).ParseFiles("templates/opds.xml")

	if err != nil {
		log.Logger.Error("error parsing OPDS template", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	// Write XML header and set content type
	w.Header().Set("Content-Type", "application/atom+xml;charset=utf-8;profile=opds-catalog;kind=acquisition")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	
	if err := tmpl.Execute(w, data); err != nil {
		log.Logger.Error("error executing OPDS template", zap.Error(err))
		// Can't send another http error response here as headers are already written
	}
}

// downloadBook handles the download of a specific book file.
func (h *Handler) downloadBook(w http.ResponseWriter, r *http.Request) {
	bookID := request.RouteIntParam(r, "id")
	if bookID == 0 {
		response.BadRequest(w, r, fmt.Errorf("invalid book ID"))
		return
	}

	book, err := h.store.GetBook(&model.FindBook{BookID: &bookID})
	if err != nil {
		log.Logger.Error("failed to get book for download", zap.Int("bookID", bookID), zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	if book == nil {
		response.NotFound(w, r)
		return
	}

	// Set header to force download with the original filename
	originalFilename := filepath.Base(book.Path)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+originalFilename+"\"")
	
	http.ServeFile(w, r, book.Path)
}

// getBaseURL determines the base URL for generating links.
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
