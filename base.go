package response

import (
	"encoding/json"
	"io"
)

// Respond creates a response with the given body.
func Respond(data []byte) Response {
	return NewResponse(func(w io.Writer) error {
		return writeAll(w, data)
	})
}

// Empty creates an empty response with the given status code.
func Empty(statusCode int) Response {
	return Respond(nil).SetStatusCode(statusCode)
}

// JSON creates a response with the data serialized as JSON for the body.
func JSON(data interface{}) Response {
	body, _ := json.Marshal(data)
	return Respond(body).SetHeader("Content-Type", "application/json")
}
