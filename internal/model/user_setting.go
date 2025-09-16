package model //import "github.com/Xunop/e-oasis/model"

import "encoding/json"

type UserSettingKey int32

const (
	UserSettingKey_USER_SETTING_KEY_UNSPECIFIED UserSettingKey = 0
	// Access tokens for the user.
	UserSettingKey_USER_SETTING_ACCESS_TOKENS UserSettingKey = 1
	// The locale of the user.
	UserSettingKey_USER_SETTING_LOCALE UserSettingKey = 2
	// The appearance of the user.
	UserSettingKey_USER_SETTING_APPEARANCE UserSettingKey = 3
	// The visibility of the memo.
	UserSettingKey_USER_SETTING_MEMO_VISIBILITY UserSettingKey = 4

	// Default view settings.
	DefaultViewSettings       = `{"show_hot_book":true}`
)

// Enum value maps for UserSettingKey.
var (
	UserSettingKey_name = map[int32]string{
		0: "USER_SETTING_KEY_UNSPECIFIED",
		1: "USER_SETTING_ACCESS_TOKENS",
		2: "USER_SETTING_LOCALE",
		3: "USER_SETTING_APPEARANCE",
		4: "USER_SETTING_MEMO_VISIBILITY",
	}
	UserSettingKey_value = map[string]int32{
		"USER_SETTING_KEY_UNSPECIFIED": 0,
		"USER_SETTING_ACCESS_TOKENS":   1,
		"USER_SETTING_LOCALE":          2,
		"USER_SETTING_APPEARANCE":      3,
		"USER_SETTING_MEMO_VISIBILITY": 4,
	}
)

type UserSetting struct {
	UserID int32
	Key    UserSettingKey
	Value  string
}

type FindUserSetting struct {
	UserID *int32
	Key    UserSettingKey
}

type ViewSetting struct {
	ShowHotBook bool `json:"show_hot_book"`
}

// AccessTokensUserSetting_AccessToken represents an access token for the user.
type AccessTokensUserSetting_AccessToken struct {
	// The access token is a JWT token.
	// Including expiration time, issuer, etc.
	AccessToken string `json:"access_token,omitempty"`
	// A description for the access token.
	Description string `json:"description,omitempty"`
	// The time when the access token was created.
	CreatedTs int64 `json:"created_ts,omitempty"`
	// The time when the access token was last used.
	LastUsedTs int64 `json:"last_used_ts,omitempty"`
}

func (a *AccessTokensUserSetting_AccessToken) String() string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

// AccessTokensUserSetting represents the access tokens for the user.
type AccessTokensUserSetting struct {
	AccessTokens []*AccessTokensUserSetting_AccessToken `json:"access_tokens,omitempty"`
}

func (a *AccessTokensUserSetting) String() string {
	if a == nil {
		return ""
	}
	b, _ := json.Marshal(a)
	return string(b)
}

func (e UserSettingKey) String() string {
	switch e {
	case UserSettingKey_USER_SETTING_ACCESS_TOKENS:
		return "USER_SETTING_ACCESS_TOKENS"
	case UserSettingKey_USER_SETTING_LOCALE:
		return "USER_SETTING_LOCALE"
	case UserSettingKey_USER_SETTING_APPEARANCE:
		return "USER_SETTING_APPEARANCE"
	case UserSettingKey_USER_SETTING_MEMO_VISIBILITY:
		return "USER_SETTING_MEMO_VISIBILITY"
	default:
		return "USER_SETTING_KEY_UNSPECIFIED"
	}
}

func (x *UserSetting) GetAccessTokens() *AccessTokensUserSetting {
	var accessTokens AccessTokensUserSetting
	if x != nil {
	    err := json.Unmarshal([]byte(x.Value), &accessTokens)
	    if err != nil {
            return nil
	    }
	    return &accessTokens
	}
	return nil
}

func (x *AccessTokensUserSetting) GetAccessTokens() []*AccessTokensUserSetting_AccessToken {
	if x != nil {
		return x.AccessTokens
	}
	return nil
}
