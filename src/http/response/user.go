package response

import (
	"github.com/Xunop/e-oasis/model"
)

func UserResponse(user *model.User) *model.User {
	return &model.User{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		ReciveBookEmail: user.ReciveBookEmail,
		Nickname:        user.Nickname,
		AvatarURL:       user.AvatarURL,
		Description:     user.Description,
		LastLoginTs:     user.LastLoginTs,
	}
}

func UserListResponse(users []*model.User) []*model.User {
	var response []*model.User
	for _, user := range users {
		response = append(response, UserResponse(user))
	}
	return response
}
