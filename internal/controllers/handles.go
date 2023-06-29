package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/presenters"
	"github.ugent.be/Universiteitsbibliotheek/handle-server-api/internal/store"
)

type Handles struct {
	Context
}

func NewHandles(c Context) *Handles {
	return &Handles{c}
}

func (handles *Handles) Get(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	localId := chi.URLParam(r, "local_id")
	handleId := prefix + "/" + localId

	var handle *store.Handle
	var pHandle *presenters.Handle
	var status int = http.StatusOK
	var hErr error

	handle, hErr = handles.Store.Get(handleId)

	if hErr != nil {
		http.Error(w, hErr.Error(), http.StatusInternalServerError)
		return
	}

	if handle == nil {
		status = http.StatusNotFound
		pHandle = presenters.EmptyResponse(handleId, 100, "handle not found")
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
	prefix := chi.URLParam(r, "prefix")
	localId := chi.URLParam(r, "local_id")
	handleId := prefix + "/" + localId

	var status int = http.StatusOK
	var responseCode int = 1
	var rowsAffected int64
	var hErr error
	rowsAffected, hErr = handles.Store.Delete(handleId)
	var message string = ""

	if hErr != nil {
		http.Error(w, hErr.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		responseCode = 100
		status = http.StatusNotFound
		message = "handle not found"
	}

	pHandle := presenters.EmptyResponse(handleId, responseCode, message)

	jsonResponse, jsonErr := json.Marshal(pHandle)

	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonResponse)
}

func (handles *Handles) Upsert(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	localId := chi.URLParam(r, "local_id")
	handleId := prefix + "/" + localId

	parseErr := r.ParseForm()

	if parseErr != nil {
		http.Error(w, parseErr.Error(), http.StatusInternalServerError)
		return
	}

	var url string = r.FormValue("url")

	if url == "" {
		http.Error(w, "url not given", http.StatusBadRequest)
		return
	}

	var handle *store.Handle = &store.Handle{
		Handle:     handleId,
		Idx:        1,
		Type:       "URL",
		Data:       url,
		Ttl:        86400,
		TtlType:    0,
		Timestamp:  time.Now().Unix(),
		AdminRead:  true,
		AdminWrite: true,
		PubRead:    true,
		PubWrite:   false,
	}

	var status int = 201
	var rowsAffected int64
	var hErr error
	rowsAffected, hErr = handles.Store.Add(handle)
	var pHandle *presenters.Handle

	if hErr != nil {
		http.Error(w, hErr.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		status = http.StatusBadRequest
		pHandle = presenters.EmptyResponse(handleId, 100, "handle not found")
	} else {
		status = 201
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
