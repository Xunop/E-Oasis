package validator // import "github.com/Xunop/e-oasis/validator"

import (
	"github.com/pkg/errors"

	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/util"
)

func ValidateUserCreateRequest(s *store.Store, user *model.UserCreateRequest) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Username == "" {
		return errors.New("username is empty")
	}
	if !util.UIDMatcher.MatchString(user.Username) {
		return errors.New("username is invalid")
	}
	if user.Email == "" {
		return errors.New("email is empty")
	}
	if user.Password == "" {
		return errors.New("password is empty")
	}
	if user, _ := s.GetUser(&model.FindUser{Username: &user.Username}); user != nil {
		return errors.New("Username already exists")
	}
	if err := validatePassword(user.Password); err != nil {
		return err
	}
	return nil
}

func ValidateSignupRequest(s *store.Store, user *model.UserSignupRequest) error {
	if user == nil {
		return errors.New("user is nil")
	}
	if user.Username == "" {
		return errors.New("username is empty")
	}
	if !util.UIDMatcher.MatchString(user.Username) {
		return errors.New("username is invalid")
	}
	if user.Password == "" {
		return errors.New("password is empty")
	}
	if user, _ := s.GetUser(&model.FindUser{Username: &user.Username}); user != nil {
		return errors.New("Username already exists")
	}
	if err := validatePassword(user.Password); err != nil {
		return err
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("password is too short")
	}
	return nil
}
