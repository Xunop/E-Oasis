package v1

import (
	"net/http"
	"os"

	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/middleware"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/worker"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
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
	sr.Use(middleware.LoggingRequest)

	// sSetting, err := store.GetSystemSecuritySetting()
	sSetting, err := store.GetOrUpsetSystemSecuritySetting()
	if err != nil {
		log.Logger.Error("Error getting security setting", zap.Error(err))
		os.Exit(1)
	}
	jwtSecret := sSetting.JWTSecret
	// Add authentication middleware
	sr.Use(NewAuthInterceptor(store, jwtSecret).AuthenticationInterceptor)
	sr.Methods(http.MethodOptions)

	sr.HandleFunc("/user", handler.createUser).Methods(http.MethodPost)
	sr.HandleFunc("/users", handler.listUsers).Methods(http.MethodGet)
    sr.HandleFunc("/signup", handler.signUp).Methods(http.MethodPost)
    sr.HandleFunc("/signin", handler.signIn).Methods(http.MethodPost)
    sr.HandleFunc("/settings/general", handler.SetGeneralSettings).Methods(http.MethodPost)
    sr.HandleFunc("/books", handler.listBooks).Methods(http.MethodGet)
    sr.HandleFunc("/book/{id}", handler.addBook).Methods(http.MethodPost)
}
