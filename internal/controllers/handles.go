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

func (h *Handles) Get(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	localId := chi.URLParam(r, "local_id")
	handleId := prefix + "/" + localId

	var handle *store.Handle
	var pHandle *presenters.Handle
	var status int = http.StatusOK
	var err error

	handle, err = h.Store.Get(r.Context(), handleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if handle == nil {
		status = http.StatusNotFound
		pHandle = presenters.EmptyResponse(handleId, 100, "handle not found")
	} else {
		status = http.StatusOK
		pHandle = presenters.FromHandle(handle)
	}

	bytes, err := json.Marshal(pHandle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(bytes)
}

func (h *Handles) Delete(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	localId := chi.URLParam(r, "local_id")
	handleId := prefix + "/" + localId

	rowsAffected, err := h.Store.Delete(r.Context(), handleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var status int = http.StatusOK
	var responseCode int = 1
	var message string = ""

	if rowsAffected == 0 {
		responseCode = 100
		status = http.StatusNotFound
		message = "handle not found"
	}

	pHandle := presenters.EmptyResponse(handleId, responseCode, message)

	bytes, err := json.Marshal(pHandle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(bytes)
}

func (h *Handles) Upsert(w http.ResponseWriter, r *http.Request) {
	prefix := chi.URLParam(r, "prefix")
	localId := chi.URLParam(r, "local_id")
	handleId := prefix + "/" + localId

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var url string = r.FormValue("url")
	if url == "" {
		http.Error(w, "url not given", http.StatusBadRequest)
		return
	}

	handle := &store.Handle{
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

	rowsAffected, err := h.Store.Add(r.Context(), handle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var status int = http.StatusCreated
	var pHandle *presenters.Handle

	if rowsAffected == 0 {
		status = http.StatusBadRequest
		pHandle = presenters.EmptyResponse(handleId, 100, "handle not found")
	} else {
		status = http.StatusCreated
		pHandle = presenters.FromHandle(handle)
	}

	bytes, err := json.Marshal(pHandle)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(bytes)
}
