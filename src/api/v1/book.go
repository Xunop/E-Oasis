package v1

import (
	"encoding/json"
	"net/http"

	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"go.uber.org/zap"
)

func (h *Handler) listBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.store.ListBooks(&model.FindBook{})
	if err != nil {
		log.Logger.Error("Error listing books", zap.Error(err))
		http.Error(w, "Error listing books", http.StatusInternalServerError)
		return
	}
	response.OK(w, r, books)
}

// addBook need to parse the format of the book and add it to the store
func (h *Handler) addBook(w http.ResponseWriter, r *http.Request) {
	book := &model.Book{}
	if err := json.NewDecoder(r.Body).Decode(book); err != nil {
		log.Logger.Error("Error decoding book", zap.Error(err))
		http.Error(w, "Error decoding book", http.StatusBadRequest)
		return
	}
	if err := h.store.AddBook(book); err != nil {
		log.Logger.Error("Error adding book", zap.Error(err))
		http.Error(w, "Error adding book", http.StatusInternalServerError)
		return
	}
	response.OK(w, r, book)
}
