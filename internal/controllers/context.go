package controllers

import (
	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/store"
)

type Context struct {
	Store  *store.Store
	Router *mux.Router
}
