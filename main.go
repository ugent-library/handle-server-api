package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ory/graceful"
	"github.com/ugent-library/zaphttp"
	"github.com/ugent-library/zaphttp/zapchi"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/controllers"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/presenters"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/store"
	"go.uber.org/zap"
)

var (
	default_options = map[string]string{
		"bind":          ":3000",
		"dsn":           "handle:handle@tcp(127.0.0.1:5432)/handle?sslmode=disable",
		"auth_username": "handle",
		"auth_password": "handle",
		"prefix":        "",
	}
	prefix        string = ""
	bind          string = ""
	dsn           string = ""
	auth_username string = ""
	auth_password string = ""
)

var logger *zap.SugaredLogger

func initLogger() {
	var l *zap.Logger
	var err error
	l, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l.Sugar()
}

func requirePrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pr := chi.URLParam(r, "prefix")
		if pr == prefix {
			next.ServeHTTP(w, r)
			return
		}

		pHandle := presenters.EmptyResponse(pr, 102, "invalid prefix")
		jsonResponse, jsonErr := json.Marshal(pHandle)

		if jsonErr != nil {
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
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

		if ok {
			if username == auth_username && password == auth_password {
				next.ServeHTTP(w, r)
				return
			}
		}

		prefix := chi.URLParam(r, "prefix")
		localId := chi.URLParam(r, "local_id")
		handleId := prefix + "/" + localId

		pHandle := presenters.EmptyResponse(handleId, 401, "not authenticated")
		jsonResponse, jsonErr := json.Marshal(pHandle)

		if jsonErr != nil {
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=handle-server-api")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		w.Write(jsonResponse)
	})
}

func main() {
	// 1. override internal default by env
	for key, _ := range default_options {
		eKey := strings.ToUpper("hdl_" + key)
		eVal := os.Getenv(eKey)
		if eVal != "" {
			default_options[key] = eVal
		}
	}

	// 2. override previous option by flag
	flag.StringVar(&prefix, "prefix", default_options["prefix"], "prefix")
	flag.StringVar(&bind, "bind", default_options["bind"], "bind")
	flag.StringVar(&dsn, "dsn", default_options["dsn"], "dsn")
	flag.StringVar(&auth_username, "auth-username", default_options["auth_username"], "basic auth username")
	flag.StringVar(&auth_password, "auth-password", default_options["auth_password"], "basic auth password")
	flag.Parse()

	if prefix == "" {
		fmt.Fprintf(os.Stderr, "Error: flag -prefix not given\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	initLogger()

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(zaphttp.SetLogger(logger.Desugar(), zapchi.RequestID))
	mux.Use(middleware.RequestLogger(zapchi.LogFormatter()))
	mux.Use(middleware.Recoverer)

	store, err := store.NewStore(dsn)
	if err != nil {
		logger.Fatal(err)
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

	srv := graceful.WithDefaults(&http.Server{
		Addr:         bind,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	logger.Info("starting handle-server-api")
	if err := graceful.Graceful(srv.ListenAndServe, srv.Shutdown); err != nil {
		logger.Fatal(err)
	}
	logger.Info("gracefully stopped server")
}
