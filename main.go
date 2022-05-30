package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/controllers"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/presenters"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/store"
)

var (
	default_options = map[string]string{
		"bind":          ":3000",
		"dsn":           "handle:handle@tcp(127.0.0.1:3306)/handle",
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

func requirePrefix(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		pr := mux.Vars(r)["prefix"]

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

		prefix := mux.Vars(r)["prefix"]
		localId := mux.Vars(r)["local_id"]
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

	router := mux.NewRouter()
	store, sErr := store.NewStore(dsn)

	if sErr != nil {
		log.Fatal(sErr.Error())
	}

	config := controllers.Context{
		Router: router,
		Store:  store,
	}

	handlesController := controllers.NewHandles(config)

	handlesRouter := router.PathPrefix("/handles").Subrouter()

	handlesRouter.Use(basicAuthMiddleware)
	handlesRouter.Use(requirePrefix)

	handlesRouter.HandleFunc("/{prefix}/{local_id}", handlesController.Get).
		Methods("GET").
		Name("get_handle")

	handlesRouter.HandleFunc("/{prefix}/{local_id}", handlesController.Delete).
		Methods("DELETE").
		Name("delete_handle")

	handlesRouter.HandleFunc("/{prefix}/{local_id}", handlesController.Upsert).
		Methods("PUT").
		Name("upsert_handle")

	log.Fatal(http.ListenAndServe(bind, handlers.LoggingHandler(os.Stdout, router)))
}
