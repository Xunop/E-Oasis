package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Xunop/e-oasis/store"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/log"
	"go.uber.org/zap"
)

type Middleware struct {
	store *store.Store
}

func NewMiddleware(store *store.Store) *Middleware {
	return &Middleware{store: store}
}

// HandleCORS handles the CORS request
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
		// FIXME: json: unsupported type: func() (io.ReadCloser, error)
		log.Debug("test in cors", zap.Any("r", r))
		next.ServeHTTP(w, r)
	})
}

// LoggingRequest logs the incomming request
func (m *Middleware) LoggingRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := request.FindClientIP(r)
		ctx := r.Context()
		ctx = context.WithValue(ctx, request.ClientIPContextKey, clientIP)

		t1 := time.Now()
		defer func() {
			log.Debug("Incomming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("proto", r.Proto),
				zap.String("client_ip", clientIP),
				zap.Duration("duration", time.Since(t1)))
		}()

		next.ServeHTTP(w, r)
	})
}
