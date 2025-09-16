package model //import "github.com/Xunop/e-oasis/model"

import (
	"testing"
)

func TestUserSettingToString(t *testing.T) {
	token := &AccessTokensUserSetting_AccessToken{
		AccessToken: "test",
		Description: "test",
		CreatedTs:   1,
		LastUsedTs:  0, // Will be ignore when marshalling
	}
	tokenString := token.String()
	if tokenString == "" {
		t.Error("Expected to get a string representation of the token")
	}
	t.Logf("Token string: %s", tokenString)
	// The tokenString should be a JSON representation of the token
	expectedJSON := `{"access_token":"test","description":"test","created_ts":1}`
	if tokenString != expectedJSON {
		t.Errorf("Expected JSON: %s, but got: %s", expectedJSON, tokenString)
	}

	accessTokens := &AccessTokensUserSetting{
		AccessTokens: []*AccessTokensUserSetting_AccessToken{token, token},
	}
	tokenStrings := accessTokens.String()
	if tokenStrings == "" {
		t.Error("Expected to get a string representation of the tokens")
	}
	expectedJSON = `{"access_tokens":[{"access_token":"test","description":"test","created_ts":1},{"access_token":"test","description":"test","created_ts":1}]}`
	if tokenStrings != expectedJSON {
		t.Errorf("Expected JSON: %s, but got: %s", expectedJSON, tokenStrings)
	}
	t.Logf("Token strings: %s", tokenStrings)
}
