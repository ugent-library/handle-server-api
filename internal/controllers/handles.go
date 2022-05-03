package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/presenters"
	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/store"
)

type Handles struct {
	Context
}

func NewHandles(c Context) *Handles {
	return &Handles{c}
}

func (handles *Handles) Get(w http.ResponseWriter, r *http.Request) {

	prefix := mux.Vars(r)["prefix"]
	localId := mux.Vars(r)["local_id"]
	handleId := prefix + "/" + localId

	var handle *store.Handle
	var pHandle *presenters.Handle
	var status int = http.StatusOK

	handle = handles.Store.Get(handleId)

	if handle == nil {

		status = http.StatusNotFound
		pHandle = presenters.EmptyResponse(handleId, 100)

	} else {

		status = http.StatusOK
		pHandle = presenters.FromHandle(handle)

	}

	jsonResponse, jsonErr := json.Marshal(pHandle)

	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

func (handles *Handles) Delete(w http.ResponseWriter, r *http.Request) {

	prefix := mux.Vars(r)["prefix"]
	localId := mux.Vars(r)["local_id"]
	handleId := prefix + "/" + localId

	var status int = http.StatusOK
	var responseCode int = 1
	var rowsAffected int64 = handles.Store.Delete(handleId)

	if rowsAffected == 0 {
		responseCode = 100
		status = http.StatusNotFound
	}

	pHandle := presenters.EmptyResponse(handleId, responseCode)

	jsonResponse, jsonErr := json.Marshal(pHandle)

	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}
