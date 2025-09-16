package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Xunop/e-oasis/api/auth"
	"github.com/Xunop/e-oasis/http/response"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/model"
	"github.com/Xunop/e-oasis/validator"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Implement the SignInWithSSO method
func (h *Handler) SignInWithSSO() {

}

func (h *Handler) signIn(w http.ResponseWriter, r *http.Request) {
	var sigin model.UserSigninRequest
	if err := json.NewDecoder(r.Body).Decode(&sigin); err != nil {
		log.Error("Failed to decode request body", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	user, err := h.store.GetUser(&model.FindUser{Username: &sigin.Username})
	if err != nil {
		log.Error("Failed to get user", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	if user == nil {
		log.Warn("User not found")
		response.NotFound(w, r)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(sigin.Password))
	if err != nil {
		log.Error("Failed to compare password", zap.Error(err))
		response.BadRequest(w, r, errors.New("Invalid password"))
		return
	}

	expireTime := time.Now().Add(auth.AccessTokenDuration)
	if sigin.NeverExpire {
		// Set the expire time to 100 years.
		expireTime = time.Now().Add(100 * 365 * 24 * time.Hour)
	}
	ctx := r.Context()
	ctx = context.WithValue(ctx, "responseWriter", w)
	ctx = context.WithValue(ctx, "request", r)
	if err := h.doSignIn(ctx, user, expireTime); err != nil {
		log.Error("Failed to sign in", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	response.OK(w, r, response.UserResponse(user))
}

func (h *Handler) doSignIn(ctx context.Context, user *model.User, expireTime time.Time) error {
	// FIXME: Why need to get secret from database every time? It should be stored in memory.
	sSetting, err := h.store.GetOrUpsetSystemSecuritySetting()
	if err != nil {
		log.Error("Failed to get security setting", zap.Error(err))
		return err
	}
	if sSetting != nil && sSetting.JWTSecret == "" {
		log.Error("JWT secret is not set")
		return errors.New("JWT secret is not set")
	}

	accessToken, err := auth.GenerateAccessToken(user.Username, user.ID, expireTime, []byte(sSetting.JWTSecret))
	if err != nil {
		log.Error("Failed to generate access token", zap.Error(err))
		return err
	}

	if err := h.store.UpsetAccessTokenToStore(user, accessToken, "User sign in"); err != nil {
		log.Error("Failed to update access token", zap.Error(err))
		return err
	}

	cookie, err := h.buildAccessTokenCookie(accessToken, expireTime, ctx.Value("request").(*http.Request).Header.Get("Origin"))
	if err != nil {
		return errors.Wrap(err, "failed to build access token cookie")
	}
	w := ctx.Value("responseWriter").(http.ResponseWriter)
	w.Header().Set("Set-Cookie", cookie)
	return nil
}

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	generalSetting, err := h.store.GetSystemGeneralSetting()
	log.Debug("General setting", zap.Any("setting", generalSetting))
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Error("Failed to get general system setting")
			response.ServerError(w, r, err)
			return
		}
	}

	// Check if signup is disabled
	if generalSetting != nil && generalSetting.DisableSignup {
		log.Debug("Signup is disabled")
		response.Forbidden(w, r)
		return
	}

	signup := &model.UserSignupRequest{}
	if err := json.NewDecoder(r.Body).Decode(&signup); err != nil {
		log.Error("Failed to decode request body", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	// Validate request
	if err := validator.ValidateSignupRequest(h.store, signup); err != nil {
		log.Error("Failed to validate signup request", zap.Error(err))
		response.BadRequest(w, r, err)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signup.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate password hash")
		response.ServerError(w, r, err)
	}
	var newRole model.Role
	hostType := model.RoleHost
	existedHostUser, err := h.store.GetUser(&model.FindUser{Role: &hostType})
	if err != nil {
		log.Error("Failed to get users", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}
	if existedHostUser == nil {
		newRole = model.RoleHost
	} else {
		newRole = model.RoleUser
	}

	user := model.User{
		Username:     signup.Username,
		Nickname:     signup.Nickname,
		PasswordHash: string(passwordHash),
		Role:         newRole,
	}

	newUser, err := h.store.CreateUser(&user)
	if err != nil {
		log.Error("Failed to signup user", zap.Error(err))
		response.ServerError(w, r, err)
		return
	}

	// Store user in cache
	h.store.UserCache.Store(newUser.ID, newUser)

	response.Created(w, r, response.UserResponse(newUser))
}

func (h *Handler) buildAccessTokenCookie(accessToken string, expireTime time.Time, origin string) (string, error) {
	attrs := []string{
		fmt.Sprintf("%s=%s", auth.AccessTokenCookieName, accessToken),
		"Path=/",
		"HttpOnly",
	}
	if expireTime.IsZero() {
		attrs = append(attrs, "Expires=Thu, 01 Jan 1970 00:00:00 GMT")
	} else {
		attrs = append(attrs, "Expires="+expireTime.Format(time.RFC1123))
	}

	if strings.HasPrefix(origin, "https://") {
		attrs = append(attrs, "Secure")
		attrs = append(attrs, "SameSite=None")
	} else {
		attrs = append(attrs, "SameSite=Lax")
	}
	return strings.Join(attrs, "; "), nil
}
