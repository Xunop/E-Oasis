package store

import (
	"strings"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Store) SetUserSetting(userSetting *model.UserSetting) error {
	query := `
		INSERT INTO user_setting (user_id, key, value)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE value = ?
	`

	log.Debug("SetUserSetting", zap.String("query", query), zap.Any("args", userSetting))

	_, err := s.db.Exec(query, userSetting.UserID, userSetting.Key.String(), userSetting.Value, userSetting.Value)
	if err != nil {
		return err
	}
	s.UserSettingCache.Store(getUserSettingCacheKey(userSetting.UserID, userSetting.Key.String()), userSetting)
	return nil
}

func (s *Store) GetUserSetting(find *model.FindUserSetting) (*model.UserSetting, error) {
	if find.UserID != nil {
		if cache, ok := s.UserSettingCache.Load(getUserSettingCacheKey(*find.UserID, find.Key.String())); ok {
			return cache.(*model.UserSetting), nil
		}
	}

	list, err := s.ListUserSettings(find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	if len(list) > 1 {
		return nil, errors.Errorf("Expected 1 user setting, but got %d", len(list))
	}

	userSetting := list[0]
	s.UserSettingCache.Store(getUserSettingCacheKey(userSetting.UserID, userSetting.Key.String()), userSetting)
	return userSetting, nil
}

func (s *Store) ListUserSettings(find *model.FindUserSetting) ([]*model.UserSetting, error) {
	where, args := []string{"1 = 1"}, []any{}

	if v := find.Key; v != model.UserSettingKey_USER_SETTING_KEY_UNSPECIFIED {
		where, args = append(where, "key = ?"), append(args, v.String())
	}
	if v := find.UserID; v != nil {
		where, args = append(where, "user_id = ?"), append(args, *find.UserID)
	}

	query := `
		SELECT
			user_id,
		  key,
			value
		FROM user_setting
		WHERE ` + strings.Join(where, " AND ")

	log.Debug("ListUserSettings", zap.String("query", query), zap.Any("args", args))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userSettingList := make([]*model.UserSetting, 0)
	for rows.Next() {
		userSetting := &model.UserSetting{}
		var keyString string
		if err := rows.Scan(
			&userSetting.UserID,
			&keyString,
			&userSetting.Value,
		); err != nil {
			return nil, err
		}
		userSetting.Key = model.UserSettingKey(model.UserSettingKey_value[keyString])
		userSettingList = append(userSettingList, userSetting)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, userSetting := range userSettingList {
		s.UserSettingCache.Store(getUserSettingCacheKey(userSetting.UserID, userSetting.Key.String()), userSetting)
	}

	return userSettingList, nil
}

// GetUserAccessTokens returns the access tokens of the user.
func (s *Store) GetUserAccessTokens(userID int32) ([]*model.AccessTokensUserSetting_AccessToken, error) {
	userSetting, err := s.GetUserSetting(&model.FindUserSetting{
		UserID: &userID,
		Key:    model.UserSettingKey_USER_SETTING_ACCESS_TOKENS,
	})
	if err != nil {
		return nil, err
	}
	if userSetting == nil {
		return []*model.AccessTokensUserSetting_AccessToken{}, nil
	}

	accessTokensUserSetting := userSetting.GetAccessTokens()
	return accessTokensUserSetting.AccessTokens, nil
}

func (s *Store) SetAPIKeyUsedTimeStamp(userID int32, token string) error {
	userTokens, err := s.GetUserAccessTokens(userID)
	if err != nil {
		return errors.Wrap(err, "unable to update last used date for API key")
	}
	if !validateAccessToken(token, userTokens) {
		return errors.New("Unable to update last used date for API key")
	}
	return nil
}

func validateAccessToken(accessTokenString string, userAccessTokens []*model.AccessTokensUserSetting_AccessToken) bool {
	for _, userAccessToken := range userAccessTokens {
		if accessTokenString == userAccessToken.AccessToken {
			return true
		}
	}
	return false
}
