package presenters

import (
	"time"

	"github.ugent.be/Universiteitsbibliotheek/hdl-srv-api/internal/store"
)

type HandleData struct {
	Url    string `json:"url"`
	Format string `json:"format"`
}

type HandleValue struct {
	Timestamp string      `json:"timestamp"`
	Type      string      `json:"type"`
	Index     int         `json:"index"`
	Ttl       int         `json:"ttl"`
	Data      *HandleData `json:"data"`
}

type Handle struct {
	Handle       string         `json:"handle"`
	ResponseCode int            `json:"responseCode"`
	Values       []*HandleValue `json:"values,omitempty"`
}

func FromHandle(h *store.Handle) *Handle {

	return &Handle{
		Handle:       h.Handle,
		ResponseCode: 1,
		Values: []*HandleValue{
			&HandleValue{
				Timestamp: time.Unix(h.Timestamp, 0).UTC().Format(time.RFC3339),
				Type:      h.Type,
				Index:     h.Idx,
				Ttl:       h.Ttl,
				Data: &HandleData{
					Url:    h.Data,
					Format: "string",
				},
			},
		},
	}

}

func EmptyResponse(h string, code int) *Handle {
	return &Handle{
		Handle:       h,
		ResponseCode: code,
	}
}
