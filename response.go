package response

import (
	"io"
	"net/http"
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

		// AddCallback registers a callback to be invoked on completion.
		AddCallback(f CallbackFunc) Response

		// WriteTo writes the response data to the ResponseWriter. This method
		// consumes the body content and will panic when called multiple times.
		WriteTo(w http.ResponseWriter)
	}

	response struct {
		statusCode int
		header     http.Header
		writer     BodyWriter
		callbacks  []CallbackFunc
		written    bool
	}

	// CallbackFunc can be registered to a response that is called after
	// the entire response body has been written to the client. If any
	// error occurred during the send, it is made available here.
	CallbackFunc func(error)

	// BodyWriter is the core of a response object - it takes a writer
	// (Which may be an http.ResponseWriter) and serializes itself to
	// it.
	BodyWriter func(io.Writer) error
)

// ensure we conform to interface
var _ Response = &response{}

// NewResponse creates a response with the given body writer.
func NewResponse(writer BodyWriter) Response {
	return &response{
		statusCode: http.StatusOK,
		header:     make(http.Header),
		writer:     writer,
	}
}

// StatusCode retrieves the status code of the response.
func (r *response) StatusCode() int {
	return r.statusCode
}

// Header retrieves the first value set to this header.
func (r *response) Header(header string) string {
	return r.header.Get(header)
}

// SetStatusCode sets the status code of the response.
func (r *response) SetStatusCode(statusCode int) Response {
	r.statusCode = statusCode
	return r
}

// SetHeader sets the value of this header.
func (r *response) SetHeader(header, val string) Response {
	r.header.Set(header, val)
	return r
}

// AddHeader adds another value to this header.
func (r *response) AddHeader(header, val string) Response {
	r.header.Add(header, val)
	return r
}

// AddCallback registers a callback to be invoked on completion.
func (r *response) AddCallback(f CallbackFunc) Response {
	r.callbacks = append(r.callbacks, f)
	return r
}

// WriteTo writes the response data to the ResponseWriter. This method
// consumes the body content and will panic when called multiple times.
func (r *response) WriteTo(w http.ResponseWriter) {
	if r.written {
		panic("response was already written")
	}

	r.written = true
	header := w.Header()

	for k, v := range r.header {
		header[k] = v
	}

	w.WriteHeader(r.statusCode)

	var err error
	if r.writer != nil {
		err = r.writer(w)
	}

	for _, c := range r.callbacks {
		c(err)
	}
}
