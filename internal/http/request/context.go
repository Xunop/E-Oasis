package request //import "github.com/Xunop/e-oasis/internal/http/request"

import (
	"net/http"

	"github.com/Xunop/e-oasis/internal/model"
	"github.com/Xunop/e-oasis/internal/log"
	"go.uber.org/zap"
)

type ContextKey int

const (
	ClientIPContextKey ContextKey = iota
	UserIDContextKey
	UserNameContextKey
	UserRolesContextKey
	IsAuthenticatedContextKey
)

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}

// ClientIP returns the client IP address stored in the context.
// func ClientIP(r *http.Request) string {
// 	return getContextStringValue(r, ClientIPContextKey)
// }

func GetUserID(r *http.Request) string {
	return getContextStringValue(r, UserIDContextKey)
}

func GetUsername(r *http.Request) string {
	return getContextStringValue(r, UserNameContextKey)
}

func GetUserRole(r *http.Request) model.Role {
	log.Info("GetUserRole", zap.Any("role", getContextStringValue(r, UserRolesContextKey)))
	return (model.Role)(getContextStringValue(r, UserRolesContextKey))
}
