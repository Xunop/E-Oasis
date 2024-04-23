package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Xunop/e-oasis/api"
	"github.com/Xunop/e-oasis/config"
	"github.com/Xunop/e-oasis/log"
	"github.com/Xunop/e-oasis/store"
	"github.com/Xunop/e-oasis/version"
	"github.com/Xunop/e-oasis/worker"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func StartServer(ctx context.Context, store *store.Store, pool *worker.Pool) (*http.Server, error) {
	addr := config.Opts.Host
	port := config.Opts.Port
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", addr, port),
		Handler: setupHandler(store, pool),
	}

	startHTTPServer(server)

	return server, nil
}

func startHTTPServer(server *http.Server) {
	go func() {
		log.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("HTTP server error", zap.Error(err))
			os.Exit(1)
		}
	}()
}

func setupHandler(store *store.Store, pool *worker.Pool) http.Handler {
	router := mux.NewRouter()

	router.Use(middleware)

	api.Server(router, store, pool)

	router.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		if err := store.Ping(); err != nil {
			http.Error(w, "Database Connection Error", http.StatusInternalServerError)
			os.Exit(1)
			return
		}

		w.Write([]byte("OK"))
	}).Name("healthcheck")

	router.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version.GetCurrentVersion()))
	}).Name("version")

	// TODO: Add metrics endpoint
	if config.Opts.MetricsCollector {
		// router.Handle("/metrics", promhttp.Handler()).Name("metrics")
		// router.Use(func(next http.Handler) http.Handler {
		// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 		route := mux.CurrentRoute(r)
		//
		// 		// Returns a 404 if the client is not authorized to access the metrics endpoint.
		// 		if route.GetName() == "metrics" && !isAllowedToAccessMetricsEndpoint(r) {
		// 			slog.Warn("Authentication failed while accessing the metrics endpoint",
		// 				slog.String("client_ip", request.ClientIP(r)),
		// 				slog.String("client_user_agent", r.UserAgent()),
		// 				slog.String("client_remote_addr", r.RemoteAddr),
		// 			)
		// 			http.NotFound(w, r)
		// 			return
		// 		}
		//
		// 		next.ServeHTTP(w, r)
		// 	})
		// })
	}

	return router
}
