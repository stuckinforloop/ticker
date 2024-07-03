package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Response struct {
	StatusCode int         `json:"status_code"`
	Data       interface{} `json:"response,omitempty"`
	Err        error       `json:"error,omitempty"`
}

func WithResponse(handler func(w http.ResponseWriter, r *http.Request) *Response) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := handler(w, r)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)

		data, err := json.MarshalIndent(response, "", " ")
		if err != nil {
			http.Error(w, "failed to marshal json response", http.StatusInternalServerError)
			return
		}

		w.Write(data)
	}
}

func (a *API) RegisterRoutes() {
	// register middlewares
	a.mux.Use(middleware.Recoverer)
	a.mux.Use(WithLogger(a.dao.Logger))

	a.mux.Route("/", func(r chi.Router) {
		r.Get("/ping", WithResponse(a.Ping))
	})

	a.mux.Route("/tasks", func(r chi.Router) {
		r.Post("/", WithResponse(a.CreateTask))
	})
}
