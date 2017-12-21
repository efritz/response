package response

import (
	"io"
	"net/http"
)

type (
	// Response wraps a payload returned from an HTTP handler function.
	Response struct {
		StatusCode int
		Header     http.Header
		writer     bodyWriter
	}

	// HandlerFunc is an analog of an http.HandlerFunc that returns a
	// response object instead of writing directly to a ResponseWriter.
	HandlerFunc func(r *http.Request) *Response

	bodyWriter func(io.Writer) error
)

// Convert converts a HandlerFunc to an http.HandlerFunc.
func Convert(f HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f(r).WriteTo(w)
	})
}

// Serialize reads the entire response and returns it as a byte slice.
// This method is meant for testing and a subsequent call to WriteTo
// will write an empty body.
func Serialize(r *Response) ([]byte, error) {
	w := newCaptureWriter(0)
	err := r.writer(w)
	return w.buffer, err
}

func newResponse(writer bodyWriter) *Response {
	return &Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		writer:     writer,
	}
}

// SetStatusCode sets the status code of the response.
func (r *Response) SetStatusCode(statusCode int) *Response {
	r.StatusCode = statusCode
	return r
}

// SetHeader sets the value of this header.
func (r *Response) SetHeader(header, val string) *Response {
	r.Header.Set(header, val)
	return r
}

// AddHeader adds another value to this header.
func (r *Response) AddHeader(header, val string) *Response {
	r.Header.Add(header, val)
	return r
}

// GetHeader retrieves the first value set to this header.
func (r *Response) GetHeader(header string) string {
	return r.Header.Get(header)
}

// WriteTo writes the response data to the ResponseWriter. This method
// consumes the underlying reader and cannot reliably be called twice.
func (r *Response) WriteTo(w http.ResponseWriter) {
	for k, v := range r.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(r.StatusCode)

	if err := r.writer(w); err != nil {
		// TODO
	}
}
