package v1

import (
	"context"
	"net/http"
	"strings"

	"github.com/Xunop/e-oasis/api/auth"
	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type AuthInterceptor struct {
	store  *store.Store
	secret string
}

func NewAuthInterceptor(store *store.Store, secret string) *AuthInterceptor {
	return &AuthInterceptor{store: store, secret: secret}
}

func (m *AuthInterceptor) AuthenticationInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isUnauthorizeAllowed(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		clientIP := request.FindClientIP(r)
		accesstoken := getAccessToken(r)

		// if accesstoken == "" {
		// 	log.Debug("Failed to authentica because no access token provided",
		// 		zap.String("client_ip", clientIP),
		// 		zap.String("user_agent", r.UserAgent()),
		// 	)
		// 	response.Unauthorized(w, r)
		// 	return
		// }

		username, err := m.authenticate(r.Context(), accesstoken)
		if err != nil {
			log.Debug("Failed to authenticate user",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.Error(err),
			)
			response.Unauthorized(w, r)
			return
		}
		user, err := m.store.GetUser(&model.FindUser{Username: &username})
		if err != nil {
			log.Error("Failed to get user",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.Error(err),
			)
			response.ServerError(w, r, err)
			return
		}
		if user == nil {
			log.Debug("User not found",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.String("username", username),
			)
			response.Unauthorized(w, r)
			return
		}
		if user.RowStatus == model.Archived {
			log.Debug("User is archived",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
				zap.String("username", username),
			)
			response.Unauthorized(w, r)
			return
		}
		if isOnlyForAdminAllowedPath(r.URL.Path) && user.Role != model.RoleHost && user.Role != model.RoleAdmin {
			response.Unauthorized(w, r)
		}

		m.store.SetLastLogin(user.ID)
		m.store.SetAPIKeyUsedTimeStamp(user.ID, accesstoken)

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.UserIDContextKey, user.ID)
		ctx = context.WithValue(ctx, request.UserNameContextKey, user.Username)
		ctx = context.WithValue(ctx, request.UserRolesContextKey, user.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthInterceptor) authenticate(ctx context.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", errors.New("no access token provided")
	}
	claims := &auth.ClaimsMessage{}
	_, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, errors.New("unexpected signing method")
		}
		if kid, ok := t.Header["kid"].(string); !ok || kid != auth.KeyID {
			return nil, errors.New("unexpected key id")
		}
		return []byte(m.secret), nil
	})
	if err != nil {
		return "", errors.New("Invalid or expired access token")
	}

	// Have valid token or generate new token
	userID, err := util.ConvertStringToInt32(claims.Subject)
	if err != nil {
		return "", errors.Wrap(err, "malformed ID in the token")
	}
	user, err := m.store.GetUser(&model.FindUser{ID: &userID})
	if err != nil {
		return "", errors.Wrap(err, "failed to get user")
	}
	if user != nil {
		return "", errors.Errorf("user not found with ID: %d", userID)
	}
	if user.RowStatus == model.Archived {
	    return "", errors.Errorf("user is archived with ID: %d", userID)
	}

    accessTokens, err := m.store.GetUserAccessTokens(userID)
    if err != nil {
        return "", errors.Wrap(err, "failed to get user access tokens")
    }

	if !validateAccessToken(accessToken, accessTokens) {
		return "", errors.New("invalid access token")
	}

	return user.Username, nil
}

func getAccessToken(r *http.Request) string {
	// Check the HTTP Authorization header first
	authorizationHeaders := r.Header.Get("Authorization")
	// Check bearer token
	if authorizationHeaders != "" {
		splitToken := strings.Split(authorizationHeaders, "Bearer ")
		if len(splitToken) == 2 {
			return splitToken[1]
		}
	}

	// Check the cookie header
	var accessToken string
	for cookie := range r.Cookies() {
		if r.Cookies()[cookie].Name == auth.AccessTokenCookieName {
			accessToken = r.Cookies()[cookie].Value
		}
	}
	return accessToken
}

func validateAccessToken(accessTokenString string, userAccessTokens []*model.AccessTokensUserSetting_AccessToken) bool {
	for _, userAccessToken := range userAccessTokens {
		if accessTokenString == userAccessToken.AccessToken {
			return true
		}
	}
	return false
}
