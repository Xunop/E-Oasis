package v1

import (
	"encoding/json"
	"net/http"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/validator"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	if request.GetUserRole(r) != model.RoleHost {
		log.Error("Unauthorized request")
		response.Unauthorized(w, r)
		return
	}

	var create model.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		log.Error("Failed to decode request body", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	// Username must use letters, numbers.
	if err := validator.ValidateUserCreateRequest(h.store, &create); err != nil {
		log.Error("Failed to validate user", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(create.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate password hash")
		response.ServerError(w, r, err)
	}
	var role model.Role
	if create.IsAdmin {
		role = model.RoleAdmin
	} else {
		role = model.RoleUser
	}
	user := model.User{
		Username:        create.Username,
		Role:            role,
		Email:           create.Email,
		ReciveBookEmail: create.ReciveBookEmail,
		Nickname:        create.Nickname,
		PasswordHash:    string(passwordHash),
		AvatarURL:       create.AvatarURL,
		Description:     create.Description,
	}

	newUser, err := h.store.CreateUser(&user)
	if err != nil {
		log.Error("Failed to create user", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	// Store user in cache
	h.store.UserCache.Store(newUser.ID, newUser)

	response.Created(w, r, response.UserResponse(newUser))
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	if request.GetUserRole(r) != model.RoleHost && request.GetUserRole(r) != model.RoleAdmin {
		log.Error("Unauthorized request by", zap.String("role", request.GetUserRole(r).String()),
			zap.String("username", request.GetUsername(r)))
		response.Unauthorized(w, r)
		return
	}

	users, err := h.store.ListUsers(&model.FindUser{})
	if err != nil {
		log.Error("Failed to list users", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	response.OK(w, r, response.UserListResponse(users))
}
