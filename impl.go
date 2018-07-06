package response

import (
	"io"
	"net/http"
)

type (
	// response implements the Response interface.
	response struct {
		statusCode int
		header     http.Header
		writer     bodyWriter
		callbacks  []CallbackFunc
		written    bool
	}

	// bodyWriter is the core of a response - it's a function that
	// takes an io.Writer (generally a response writer) and serializes
	// the response body to it.
	bodyWriter func(io.Writer) error
)

// ensure we conform to interface
var _ Response = &response{}

// newResponse creates a response with the given body writer.
func newResponse(writer bodyWriter) Response {
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
	if val == "" {
		r.header.Del(header)
	} else {
		r.header.Set(header, val)
	}

	return r
}

// AddHeader adds another value to this header.
func (r *response) AddHeader(header, val string) Response {
	r.header.Add(header, val)
	return r
}

// AddCallback registers a callback to be invoked on after the entire
// response body has been written to the client. If any error occurred
// during the send, it is made available to the function registered here.
func (r *response) AddCallback(f CallbackFunc) Response {
	r.callbacks = append(r.callbacks, f)
	return r
}

// DecorateWriter wraps a function around the underlying io.Writer which
// writes the response body content. This method will wrap the decorated
// writer with a CloseNotify method which is delegated from the edge
// response writer. Once the body writer is evaluated to completion, the
// decorated writer is closed.
func (r *response) DecorateWriter(f WriterDecorator) Response {
	baseWriter := r.writer

	r.writer = func(w io.Writer) error {
		decorated := f(w)
		err := baseWriter(delegateCloseNotify(decorated, w))
		return tryClose(decorated, err)
	}

	return r
}

// tryClose attempts to close the given writer. The original error is
// returned if non-nil. Otherwise, the error result from the Close method
// is returned.
func tryClose(w io.Writer, originalErr error) error {
	if c, ok := w.(io.Closer); ok {
		if err := c.Close(); originalErr == nil {
			return err
		}
	}

	return originalErr
}

// delegateCloseNotify bundles a CloseNotify method with writer w1 if w2
// is a close notifier. The channel returned by the new method will close
// when the close notification channel of writer w2 closes.
func delegateCloseNotify(w1, w2 io.Writer) io.Writer {
	if cn, ok := w2.(http.CloseNotifier); ok {
		ch := make(chan bool)

		go func() {
			<-cn.CloseNotify()
			close(ch)
		}()

		return &closeableWriter{w1, ch}
	}

	return w1
}

// WriteTo writes the response data to the ResponseWriter. This method
// consumes the body content and will panic when called multiple times.
func (r *response) WriteTo(w http.ResponseWriter) {
	if r.written {
		panic("response was already written")
	}

	r.written = true
	r.writeHeader(w)
	err := r.writeBody(w)

	for _, c := range r.callbacks {
		c(err)
	}
}

// writeHeader writes the headers and status code to the response writer.
func (r *response) writeHeader(w http.ResponseWriter) {
	header := w.Header()
	for k, v := range r.header {
		header[k] = v
	}

	w.WriteHeader(r.statusCode)
}

// writeBody writes the entire body to the response writer (if any writer
// is supplied).
func (r *response) writeBody(w http.ResponseWriter) error {
	if r.writer == nil {
		return nil
	}

	return r.writer(w)
}
