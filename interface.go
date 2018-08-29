package response

import (
	"io"
	"net/http"
	"net/http/httptest"
)

type (
	// Response wraps a payload returned from an HTTP handler function.
	Response interface {
		// StatusCode retrieves the status code of the response.
		StatusCode() int

		// Header retrieves the first value set to this header.
		Header(header string) string

		// SetStatusCode sets the status code of the response.
		SetStatusCode(statusCode int) Response

		// SetHeader sets the value of this header.
		SetHeader(header, val string) Response

		// AddHeader adds another value to this header.
		AddHeader(header, val string) Response

		// AddCallback registers a callback to be invoked on after
		// the entire response body has been written to the client.
		AddCallback(f CallbackFunc) Response

		// DecorateWriter wraps a function around the underlying io.Writer
		// which writes the response body content.
		DecorateWriter(f WriterDecorator) Response

		// WriteTo writes the response data to the ResponseWriter. This method
		// consumes the body content and will panic when called multiple times.
		WriteTo(w http.ResponseWriter)
	}

	// CallbackFunc receives the value of errors which occur when the body
	// fails to write to the remote end.
	CallbackFunc func(error)

	// WriterDecorator returns an io.Writer that writes to the given io.Writer.
	WriterDecorator func(io.Writer) io.Writer

	// HandlerFunc is an analog of an http.HandlerFunc that returns a
	// response object instead of writing directly to a ResponseWriter.
	HandlerFunc func(*http.Request) Response
)

// Serialize reads the entire response and returns the headers and a
// byte slice containing the content of the entire body. An error is
// returned if writing to the response recorder fails.
func Serialize(r Response) (http.Header, []byte, error) {
	w := httptest.NewRecorder()

	var err error
	r.AddCallback(func(e error) { err = e })
	r.WriteTo(w)

	return w.HeaderMap, w.Body.Bytes(), err
}

// Reconstruct creates a response from the values returned from Serialize.
func Reconstruct(statusCode int, headers http.Header, body []byte) Response {
	resp := Respond(body)
	resp.SetStatusCode(statusCode)

	for k, vs := range headers {
		resp.SetHeader(k, vs[0])

		for _, v := range vs[1:] {
			resp.AddHeader(k, v)
		}
	}

	return resp
}

// Convert converts a HandlerFunc to an http.HandlerFunc.
func Convert(f HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f(r).WriteTo(w)
	})
}
