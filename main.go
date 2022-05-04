package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/controllers"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/store"
)

var bind string = ":3000"
var dsn string = "handle:handle@tcp(127.0.0.1:3306)/handle"
var auth_username = "handle"
var auth_password = "handle"

func basicAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()

		if ok {
			if username == auth_username && password == auth_password {

				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", "Basic realm=hdl-srv-api")
		w.WriteHeader(401)
		w.Write([]byte("Unauthorised.\n"))
	})
}

func main() {

	flag.StringVar(&bind, "bind", bind, "bind")
	flag.StringVar(&dsn, "dsn", dsn, "dsn")
	flag.StringVar(&auth_username, "auth-username", auth_username, "basic auth username")
	flag.StringVar(&auth_password, "auth-password", auth_password, "basic auth password")

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

	router.Use(basicAuthMiddleware)

	handlesController := controllers.NewHandles(config)

	handlesRouter := router.PathPrefix("/handles").Subrouter()
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
