package v1

import (
	"fmt"
	"html/template"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// OpdsEntry can be either a link to another feed (navigation) or a book (acquisition).
type OpdsEntry struct {
	ID       string
	Title    string
	Author   string
	Content  string
	Updated  time.Time
	IsNav    bool // True if this is a link to another feed
	NavURL   string
	AcqURL   string // Acquisition (download) URL
	MimeType string
	CoverURL string
	HasCover bool
}

// OpdsTemplateData now holds entries that can be navigation or acquisition.
type OpdsTemplateData struct {
	ID             string
	Title          string
	Entries        []*OpdsEntry
	BaseURL        string
	CurrentTime    string
	RequestURLPath string
}

// OpdsRootFeed is the new main entry point at /opds.
func (h *Handler) opdsRootFeed(w http.ResponseWriter, r *http.Request) {
	baseURL := getBaseURL(r)
	data := OpdsTemplateData{
		ID:          fmt.Sprintf("%s/opds", baseURL),
		Title:       "E-Oasis Library",
		BaseURL:     baseURL,
		CurrentTime: time.Now().UTC().Format(time.RFC3339),
		Entries: []*OpdsEntry{
			{
				ID:      fmt.Sprintf("%s/opds/all", baseURL),
				Title:   "All Books",
				Content: "Complete list of all books",
				Updated: time.Now().UTC(),
				IsNav:   true,
				NavURL:  fmt.Sprintf("%s/opds/all", baseURL),
			},
			{
				ID:      fmt.Sprintf("%s/opds/tags", baseURL),
				Title:   "Browse by Tag",
				Content: "Browse books sorted by tag/genre",
				Updated: time.Now().UTC(),
				IsNav:   true,
				NavURL:  fmt.Sprintf("%s/opds/tags", baseURL),
			},
		},
		RequestURLPath: r.URL.Path,
	}
	h.renderOpdsTemplate(w, r, data)
}

// OpdsAllBooksFeed handles the flat list of all books.
func (h *Handler) opdsAllBooksFeed(w http.ResponseWriter, r *http.Request) {
	// Fetch all books from the store
	books, err := h.store.ListBooks(&model.FindBook{})
	if err != nil {
		log.Logger.Error("failed to list books for OPDS 'all' feed", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	// Pass the list of books to the reusable helper function to render the feed.
	h.serveAcquisitionFeed(w, r, "All Books", books)
}

// OpdsTagsFeed lists all available tags.
func (h *Handler) opdsTagsFeed(w http.ResponseWriter, r *http.Request) {
	tags, err := h.store.ListAllTags()
	if err != nil {
		log.Logger.Error("failed to list books tags", zap.Error(err))
		return
	}

	baseURL := getBaseURL(r)
	entries := make([]*OpdsEntry, len(tags))
	for i, tag := range tags {
		entries[i] = &OpdsEntry{
			ID:      fmt.Sprintf("%s/opds/tags/%d", baseURL, tag.ID),
			Title:   tag.Name,
			Content: fmt.Sprintf("%d books", tag.BookCount),
			Updated: time.Now().UTC(),
			IsNav:   true,
			NavURL:  fmt.Sprintf("%s/opds/tags/%d", baseURL, tag.ID),
		}
	}

	data := OpdsTemplateData{
		ID:             fmt.Sprintf("%s/opds/tags", baseURL),
		Title:          "Browse by Tag",
		BaseURL:        baseURL,
		CurrentTime:    time.Now().UTC().Format(time.RFC3339),
		Entries:        entries,
		RequestURLPath: r.URL.Path,
	}
	h.renderOpdsTemplate(w, r, data)
}

// OpdsBooksByTagFeed lists books for a specific tag.
func (h *Handler) opdsBooksByTagFeed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tagID, _ := strconv.Atoi(vars["id"])

	books, err := h.store.ListBooksByTag(tagID)
	if err != nil {
		log.Logger.Error("failed to list books tags", zap.Error(err))
		return
	}

	// You might want a store method to get tag name from ID for the title
	h.serveAcquisitionFeed(w, r, fmt.Sprintf("Tag ID: %d", tagID), books)
}

// serveAcquisitionFeed is a helper to render a list of books.
func (h *Handler) serveAcquisitionFeed(w http.ResponseWriter, r *http.Request, title string, books []*model.Book) {
	baseURL := getBaseURL(r)
	entries := make([]*OpdsEntry, len(books))

	// It converts each model.Book into an OpdsEntry.
	for i, book := range books {
		lastModifiedTime, _ := time.Parse(time.RFC3339, book.LastModified)

		// Get MIME type from extension
		mimeType := mime.TypeByExtension(filepath.Ext(book.Path))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		entries[i] = &OpdsEntry{
			ID:       fmt.Sprintf("urn:uuid:%s", book.UUID),
			Title:    book.Title,
			Author:   book.AuthorSort,
			Updated:  lastModifiedTime,
			IsNav:    false, // This is an acquisition feed, so IsNav is always false.
			AcqURL:   fmt.Sprintf("%s/opds/download/%d", baseURL, book.ID),
			MimeType: mimeType,
			HasCover: book.HasCover,
			CoverURL: fmt.Sprintf("%s/api/v1/covers/%d", baseURL, book.ID),
		}
	}

	// Prepare the data for the template, using the title and entries we just created.
	data := OpdsTemplateData{
		ID:             fmt.Sprintf("%s%s", baseURL, r.URL.Path),
		Title:          title,
		BaseURL:        baseURL,
		CurrentTime:    time.Now().UTC().Format(time.RFC3339),
		Entries:        entries,
		RequestURLPath: r.URL.Path,
	}

	// Call the final rendering helper.
	h.renderOpdsTemplate(w, r, data)
}

// renderOpdsTemplate is a centralized helper to render any OPDS feed.
func (h *Handler) renderOpdsTemplate(w http.ResponseWriter, r *http.Request, data OpdsTemplateData) {
	tmpl, err := template.New("opds.xml").Funcs(template.FuncMap{
		"formatDate": func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
	}).ParseFiles("templates/opds.xml")
	if err != nil {
		log.Logger.Error("error parse OPDS template", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/atom+xml;charset=utf-8;profile=opds-catalog;kind=navigation")
	fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`)
	tmpl.Execute(w, data)
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
