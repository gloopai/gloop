package site

import (
	"encoding/json"
	"net/http"
)

type RequestAuth struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
}

type RequestPayload struct {
	Auth    RequestAuth `json:"auth"`
	Command string      `json:"command"`
	Data    interface{} `json:"data"`
}

type ResponsePayload struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ParseJSONRequest(r *http.Request, payload *RequestPayload) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(payload)
}

func WriteJSONResponse(w http.ResponseWriter, response ResponsePayload) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
