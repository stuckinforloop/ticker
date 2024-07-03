package api

import "net/http"

func (a *API) Ping(w http.ResponseWriter, r *http.Request) *Response {
	return &Response{
		StatusCode: http.StatusOK,
		Data:       "pong!",
	}
}
