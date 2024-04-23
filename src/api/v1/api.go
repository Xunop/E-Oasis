package v1

import (
	"github.com/Xunop/e-oasis/middleware"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/worker"
	"github.com/gorilla/mux"
)

type Handler struct {
	store  *store.Store
	pool   *worker.Pool
	router *mux.Router
}

func Server(router *mux.Router, store *store.Store, pool *worker.Pool) {
	handler := &Handler{
		store:  store,
		pool:   pool,
		router: router,
	}

	sr := router.PathPrefix("/api/v1").Subrouter()
	middleware := middleware.NewMiddleware(handler.store)
	sr.Use(middleware.HandleCORS)
}
