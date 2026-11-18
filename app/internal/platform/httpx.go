package httpx

import (
	"encoding/json"
	"net/http"
)

type ResponseEnvelope struct {
	Data any `json:"data"`
}

// Represents an error structure as described in RFC9457
//
// Type is an URI representing the error type which can be the text value of the http status code
// Status is the numeric HTTP responswe code
// Title is a short human readable description about the error
// Detail represents the detailed error, can be a map or any structure that can be represented by json
//
// More info: https://www.rfc-editor.org/info/rfc9457/
type ProblemDetail struct {
	Type     string `json:"type"`
	Status   int    `json:"status"`
	Title    string `json:"title"`
	Detail   any    `json:"detail"`
	Instance string `json:"instance"`
}

func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, r *http.Request, status int, message string, detail any) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)

	response := ProblemDetail{
		Type:     http.StatusText(status),
		Status:   status,
		Title:    message,
		Detail:   detail,
		Instance: r.URL.RequestURI(),
	}
	json.NewEncoder(w).Encode(response)
}

func JSON(w http.ResponseWriter, status int, data any) {
	writeJson(w, status, &ResponseEnvelope{Data: data})
}
