package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Xunop/e-oasis/http/request"
	"github.com/Xunop/e-oasis/log"
	"go.uber.org/zap"
)

func middleware(next http.Handler) http.Handler {
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
