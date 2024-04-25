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

		if cache, ok := s.userCache.Load(*find.ID); ok {
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
	s.userCache.Store(user.ID, user)
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

	orderBy := []string{"created_ts DESC", "row_status DESC"}
	if find.Random {
		orderBy = slices.Concat([]string{"RANDOM()"}, orderBy)
	}

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
			row_status
		FROM user
		WHERE ` + strings.Join(where, " AND ") + ` ORDER BY ` + strings.Join(orderBy, ", ")
	if v := find.Limit; v != nil {
		query += fmt.Sprintf(" LIMIT %d", *v)
	}

	log.Debug("ListUsers", zap.String("query", query), zap.Any("args", args))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]*model.User, 0)
	for rows.Next() {
		var user model.User
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
	_, err := s.db.Exec(query, userID)
	if err != nil {
		errors.Wrap(err, "store: unable to update last login date")
	}
	return nil
}
