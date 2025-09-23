package v1

import (
	"net/http"

	"github.com/Xunop/e-oasis/internal/http/request"
	"github.com/Xunop/e-oasis/internal/model"
	"github.com/Xunop/e-oasis/internal/store"
)

func getCurrentUser(r *http.Request, s *store.Store) (*model.User, error) {
	// Get the current user from the request
	username := request.GetUsername(r)
	if username == "" {
		// If the user is not logged in, return an error
		return nil, nil
	}

	// Get the user from the store
	user, err := s.GetUser(&model.FindUser{Username: &username})
	if err != nil {
		// If an error occurs, return an error
		return nil, err
	}
	return user, nil
}
