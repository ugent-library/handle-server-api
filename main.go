package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/controllers"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/presenters"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/store"
)

var (
	default_options = map[string]string{
		"bind":          ":3000",
		"dsn":           "handle:handle@tcp(127.0.0.1:3306)/handle",
		"auth_username": "handle",
		"auth_password": "handle",
	}
	bind          string = ""
	dsn           string = ""
	auth_username        = ""
	auth_password        = ""
)

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

		w.Header().Set("WWW-Authenticate", "Basic realm=hdl-srv-api")
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
	flag.StringVar(&bind, "bind", default_options["bind"], "bind")
	flag.StringVar(&dsn, "dsn", default_options["dsn"], "dsn")
	flag.StringVar(&auth_username, "auth-username", default_options["auth_username"], "basic auth username")
	flag.StringVar(&auth_password, "auth-password", default_options["auth_password"], "basic auth password")

	flag.Parse()

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
