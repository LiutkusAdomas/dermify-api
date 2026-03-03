package handlers

import (
	"dermify-api/internal/api/metrics"
	"encoding/json"
	"net/http"
)

type fooResponse struct {
	Foo string `json:"foo" example:"Bar"`
}

// HandleFoo returns a foo bar response and increments the foo counter.
//
//	@Summary		Foo bar
//	@Description	Returns a foo bar response and increments the foo metric counter
//	@Tags			general
//	@Produce		json
//	@Success		200	{object}	fooResponse
//	@Router			/foo [get]
func HandleFoo(metrics *metrics.Client) func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		metrics.IncrementFooCount()
		res := fooResponse{Foo: "Bar"}
		b, _ := json.Marshal(res)
		_, _ = w.Write(b)
	}
}
