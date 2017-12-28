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
		callbacks  []CallbackFunc
		written    bool
	}

	// HandlerFunc is an analog of an http.HandlerFunc that returns a
	// response object instead of writing directly to a ResponseWriter.
	HandlerFunc func(*http.Request) *Response

	// CallbackFunc can be registered to a response that is called after
	// the entire response body has been written to the client. If any
	// error occurred during the send, it is made available here.
	CallbackFunc func(error)

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

// AddCallback registers a callback to be invoked on completion.
func (r *Response) AddCallback(f CallbackFunc) *Response {
	r.callbacks = append(r.callbacks, f)
	return r
}

// GetHeader retrieves the first value set to this header.
func (r *Response) GetHeader(header string) string {
	return r.Header.Get(header)
}

// WriteTo writes the response data to the ResponseWriter. This method
// consumes the body content and will panic when called multiple times.
func (r *Response) WriteTo(w http.ResponseWriter) {
	if r.written {
		panic("response was already written")
	}

	r.written = true
	header := w.Header()

	for k, v := range r.Header {
		header[k] = v
	}

	w.WriteHeader(r.StatusCode)
	err := r.writer(w)

	for _, c := range r.callbacks {
		c(err)
	}
}
