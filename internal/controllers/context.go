package controllers

import (
	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/store"
)

type Context struct {
	Store  *store.Store
	Router *mux.Router
}
