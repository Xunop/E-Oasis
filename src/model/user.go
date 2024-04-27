package model

// Role is the type of a role.
type Role string

const (
	// RoleHost is the HOST role.
	RoleHost Role = "HOST"
	// RoleAdmin is the ADMIN role.
	RoleAdmin Role = "ADMIN"
	// RoleUser is the USER role.
	RoleUser Role = "USER"
)

func (e Role) String() string {
	switch e {
	case RoleHost:
		return "HOST"
	case RoleAdmin:
		return "ADMIN"
	case RoleUser:
		return "USER"
	}
	return "USER"
}

const (
	SystemBotID         int32 = 0
)

var (
	SystemBot = &User{
		ID:       SystemBotID,
		Username: "system_bot",
		Role:     RoleAdmin,
		Email:    "",
		Nickname: "Bot",
	}
)

type User struct {
	ID int32 `json:"id"`

	RowStatus RowStatus `json:"row_status"`
	CreatedTs int64     `json:"created_ts"`
	UpdatedTs int64     `json:"updated_ts"`

	Username        string `json:"username"`
	Role            Role   `json:"role"`
	Email           string `json:"email"`
	ReciveBookEmail string `json:"recive_book_email"`
	Nickname        string `json:"nickname"`
	PasswordHash    string `json:"password_hash"`
	AvatarURL       string `json:"avatar_url"`
	Description     string `json:"description"`
	LastLoginTs     int64  `json:"last_login_ts"`
}

type FindUser struct {
	ID        *int32     `json:"id"`
	RowStatus *RowStatus `json:"row_status"`
	Username  *string    `json:"username"`
	Role      *Role      `json:"role"`
	Email     *string    `json:"email"`
	Nickname  *string    `json:"nickname"`

	// Random and limit are used in list users.
	// Whether to return random users.
	Random bool
	// The maximum number of users to return.
	Limit *int
}

type UserCreateRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	IsAdmin         bool   `json:"is_admin"`
	Email           string `json:"email"`
	ViewSettings    string `json:"view_settings"`
	ReciveBookEmail string   `json:"recive_book_email"`
	Nickname        string `json:"nickname"`
	AvatarURL       string `json:"avatar_url"`
	Description     string `json:"description"`
}

type UserSignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}
