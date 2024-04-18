package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ory/graceful"
	"github.com/spf13/cobra"
	"github.com/ugent-library/zaphttp"
	"github.com/ugent-library/zaphttp/zapchi"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/controllers"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/presenters"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/store"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

func requirePrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := chi.URLParam(r, "prefix")
		if prefix == config.Prefix {
			next.ServeHTTP(w, r)
			return
		}

		pHandle := presenters.EmptyResponse(prefix, 102, "invalid prefix")
		jsonResponse, err := json.Marshal(pHandle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(jsonResponse)
	})
}

func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if ok && username == config.AuthUsername && password == config.AuthPassword {
			next.ServeHTTP(w, r)
			return
		}

		prefix := chi.URLParam(r, "prefix")
		localId := chi.URLParam(r, "local_id")
		handleId := prefix + "/" + localId
		pHandle := presenters.EmptyResponse(handleId, 401, "not authenticated")
		jsonResponse, err := json.Marshal(pHandle)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=handle-server-api")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write(jsonResponse)
	})
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start the handle server",
	RunE: func(cmd *cobra.Command, args []string) error {
		mux := chi.NewRouter()
		mux.Use(middleware.RequestID)
		mux.Use(middleware.RealIP)
		mux.Use(zaphttp.SetLogger(logger.Desugar(), zapchi.RequestID))
		mux.Use(middleware.RequestLogger(zapchi.LogFormatter()))
		mux.Use(middleware.Recoverer)

		store, err := store.NewStore(config.DSN)
		if err != nil {
			return err
		}

		handlesController := controllers.NewHandles(controllers.Context{
			Store: store,
		})

		mux.Route("/handles", func(r chi.Router) {
			r.With(basicAuthMiddleware, requirePrefix).Route("/{prefix}/{local_id}", func(r chi.Router) {
				r.Get("/", handlesController.Get)
				r.Delete("/", handlesController.Delete)
				r.Put("/", handlesController.Upsert)
			})
		})
		mux.Get("/info", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			bytes, _ := json.Marshal(&struct {
				Branch string `json:"branch,omitempty"`
				Commit string `json:"commit,omitempty"`
				Image  string `json:"image,omitempty"`
			}{
				Branch: config.Version.Branch,
				Commit: config.Version.Commit,
				Image:  config.Version.Image,
			})
			w.Write(bytes)
		})
		mux.Get("/status", health.NewHandler(health.NewChecker(
			health.WithCheck(health.Check{
				Name:    "database",
				Timeout: 5 * time.Second,
				Check:   store.Ping,
			}),
		)))

		srv := graceful.WithDefaults(&http.Server{
			Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		})

		logger.Infof("starting server at %s:%d", config.Host, config.Port)
		if err := graceful.Graceful(srv.ListenAndServe, srv.Shutdown); err != nil {
			return err
		}
		logger.Info("gracefully stopped server")

		return nil
	},
}
