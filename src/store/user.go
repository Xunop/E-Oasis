package store

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *Store) GetUser(find *model.FindUser) (*model.User, error) {
	if find.ID != nil {
		if *find.ID == model.SystemBotID {
			return model.SystemBot, nil
		}

		if cache, ok := s.UserCache.Load(*find.ID); ok {
			return cache.(*model.User), nil
		}
	}

	list, err := s.ListUsers(find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}

	user := list[0]
	s.UserCache.Store(user.ID, user)
	return user, nil
}

func (s *Store) ListUsers(find *model.FindUser) ([]*model.User, error) {
	where, args := []string{"1 = 1"}, []any{}

	if v := find.ID; v != nil {
		where, args = append(where, "id = ?"), append(args, *v)
	}
	if v := find.Username; v != nil {
		where, args = append(where, "username = ?"), append(args, *v)
	}
	if v := find.Role; v != nil {
		where, args = append(where, "role = ?"), append(args, *v)
	}
	if v := find.Email; v != nil {
		where, args = append(where, "email = ?"), append(args, *v)
	}
	if v := find.Nickname; v != nil {
		where, args = append(where, "nickname = ?"), append(args, *v)
	}

	// Get only active users
	orderBy := []string{"created_ts DESC", "row_status DESC"}
	if find.Random {
		orderBy = slices.Concat([]string{"RANDOM()"}, orderBy)
	}

	// Here will return password_hash, so need to be careful
	// If need to response to client, need to remove password_hash
	// Use response.UserResponse to remove password_hash
	query := `
		SELECT
			id,
			username,
			role,
			email,
			nickname,
			password_hash,
			avatar_url,
			description,
			created_ts,
			updated_ts,
            last_login_ts,
			row_status,
	        recive_book_email
		FROM user
		WHERE ` + strings.Join(where, " AND ") + ` ORDER BY ` + strings.Join(orderBy, ", ")
	if v := find.Limit; v != nil {
		query += fmt.Sprintf(" LIMIT %d", *v)
	}

	// zap not support escape character, so need to fallback.
	// https://github.com/uber-go/zap/issues/963
	log.Debug("SQL query and args:")
	log.Fallback("Debug", fmt.Sprintf("query: %s\nargs: %s\n", query, args))

	rows, err := s.appDb.Query(query, args...)
	if err != nil {
		log.Debug("Error querying users", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.User, 0)
	for rows.Next() {
		var user model.User
		// The ordering of query results should be consistent with query var
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Role,
			&user.Email,
			&user.Nickname,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.Description,
			&user.CreatedTs,
			&user.UpdatedTs,
			&user.LastLoginTs,
			&user.RowStatus,
			&user.ReciveBookEmail,
		); err != nil {
			return nil, err
		}
		list = append(list, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (s *Store) SetLastLogin(userID int32) error {
	query := `UPDATE user SET last_login_ts=now() WHERE id=$1`
	_, err := s.appDb.Exec(query, userID)
	if err != nil {
		errors.Wrap(err, "store: unable to update last login date")
	}
	return nil
}

func (s *Store) CreateUser(create *model.User) (*model.User, error) {
	fields := []string{"`username`", "`role`", "`email`", "`recive_book_email`", "`nickname`", "`password_hash`", "`avatar_url`", "`description`"}
	placeholder := []string{"?", "?", "?", "?", "?", "?", "?", "?"}
	args := []any{create.Username, create.Role, create.Email, create.ReciveBookEmail, create.Nickname, create.PasswordHash, create.AvatarURL, create.Description}
	stmt := "INSERT INTO user (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholder, ", ") + ") RETURNING id, row_status, created_ts, updated_ts, last_login_ts, username, role, email, recive_book_email, nickname, avatar_url, description"

	// log.Debug("CreateUser", zap.String("stmt", stmt), zap.Any("args", args))
	log.Fallback("Debug", fmt.Sprintf("CreateUser\nstmt: %s\nargs: %s\n", stmt, args))

	tx, err := s.appDb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user model.User
	if err := tx.QueryRow(stmt, args...).Scan(
		&user.ID,
		&user.RowStatus,
		&user.CreatedTs,
		&user.UpdatedTs,
		&user.LastLoginTs,
		&user.Username,
		&user.Role,
		&user.Email,
		&user.ReciveBookEmail,
		&user.Nickname,
		&user.AvatarURL,
		&user.Description,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &user, nil
}
