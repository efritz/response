package response

import (
	"encoding/json"
	"fmt"
	"io"
)

// Respond creates a response with the given body.
func Respond(data []byte) Response {
	writer := func(w io.Writer) error {
		return writeAll(w, data)
	}

	resp := newResponse(writer)
	resp.SetHeader("Content-Length", fmt.Sprintf("%d", len(data)))
	return resp
}

// Empty creates an empty response with the given status code.
func Empty(statusCode int) Response {
	resp := Respond(nil)
	resp.SetStatusCode(statusCode)
	resp.SetHeader("Content-Length", "0")
	return resp
}

// JSON creates a response with the data serialized as JSON for the body.
func JSON(data interface{}) Response {
	body, _ := json.Marshal(data)

	resp := Respond(body)
	resp.SetHeader("Content-Type", "application/json")
	resp.SetHeader("Content-Length", fmt.Sprintf("%d", len(body)))
	return resp
}
