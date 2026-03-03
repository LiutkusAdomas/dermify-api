package handlers

import (
	"encoding/json"
	"net/http"
)

type helloResponse struct {
	Hello string `json:"hello" example:"World"`
}

// HandleHello returns a hello world response.
//
//	@Summary		Hello world
//	@Description	Returns a simple hello world message
//	@Tags			general
//	@Produce		json
//	@Success		200	{object}	helloResponse
//	@Router			/hello [get]
func HandleHello() func(w http.ResponseWriter, _ *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		res := helloResponse{Hello: "World"}
		b, _ := json.Marshal(res)
		_, _ = w.Write(b)
	}
}
