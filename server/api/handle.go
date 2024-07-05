package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/viper"
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

	allowedOrigins := make([]string, 0)
	env := viper.GetString("environment")
	switch env {
	case "dev":
		allowedOrigins = append(allowedOrigins, "http://localhost:3000")
	case "prod":
		// FIXME: add valid origins
	}

	a.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	a.mux.Route("/", func(r chi.Router) {
		r.Get("/ping", WithResponse(a.Ping))
	})

	a.mux.Route("/tasks", func(r chi.Router) {
		r.Get("/", WithResponse(a.GetTasks))
		r.Post("/", WithResponse(a.CreateTask))

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", WithResponse(a.GetTask))
			r.Patch("/", WithResponse(a.UpdateTask))
			r.Delete("/", WithResponse(a.DeleteTask))
		})
	})
}
