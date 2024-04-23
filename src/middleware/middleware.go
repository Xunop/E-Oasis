package middleware

import (
	"net/http"

	"github.com/Xunop/e-oasis/store"
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
