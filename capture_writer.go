package response

import (
	"net/http"
)

// CaptureWriter is an http.ResponseWriter that saves all of the data to
// memory to be referenced later. This can be used in place of a standard
// http.ResponseWriter for testing purposes (as used in this library), or
// to convert or adapt an HTTP response in-flight.
type CaptureWriter struct {
	numWrites  int
	StatusCode int
	Body       []byte
	header     http.Header
	chunkSize  int
}

// NewCaptureWriter creates a new CaptureWriter with the given read
// chunk size. Supplying a chunkSize of zero makes the writes unbounded.
func NewCaptureWriter(chunkSize int) *CaptureWriter {
	return &CaptureWriter{
		chunkSize: chunkSize,
		Body:      []byte{},
		header:    http.Header{},
	}
}

// WriteHeader captures the HTTP status code.
func (w *CaptureWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

// Header returns the captured header map.
func (w *CaptureWriter) Header() http.Header {
	return w.header
}

// Write appends this data to the captured HTTP body.
func (w *CaptureWriter) Write(p []byte) (int, error) {
	if w.chunkSize < len(p) && w.chunkSize > 0 {
		p = p[:w.chunkSize]
	}

	w.numWrites++
	w.Body = append(w.Body, p...)
	return len(p), nil
}
