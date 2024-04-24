package middleware

import (
	"net/http"
	"strings"

	"github.com/Xunop/e-oasis/api/auth"
	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/store"
	"go.uber.org/zap"
)

type Middleware struct {
	store *store.Store
}

func NewMiddleware(store *store.Store) *Middleware {
	return &Middleware{store: store}
}

func (m *Middleware) HandleCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Auth-Token, Authorization, Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "7200")
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) AuthenticationInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.ClientIP(r)
		accesstoken := getAccessToken(r)

		if accesstoken == "" {
			log.Debug("[API] Skipping authentication because no access token provided",
				zap.String("client_ip", clientIP),
				zap.String("user_agent", r.UserAgent()),
			)
			// json.Unauthorized(w, "No access token provided")
		}

	})
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
