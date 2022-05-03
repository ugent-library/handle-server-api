package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/controllers"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/store"
)

func basicAuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		username, password, ok := r.BasicAuth()

		if ok {
			if username == "handle" && password == "handle" {

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

	router := mux.NewRouter()
	store, sErr := store.NewStore("handle:handle@tcp(127.0.0.1:3307)/handle")

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

	log.Fatal(http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, router)))
}
