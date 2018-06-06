package response

import (
	"errors"
	"io"
	"net/http"
)

type closer struct {
	io.Reader
	closed bool
}

func (c *closer) Read(buffer []byte) (int, error) {
	if c.closed {
		return 0, io.EOF
	}

	return c.Reader.Read(buffer)
}

func (c *closer) Close() error {
	c.closed = true
	return nil
}

//
//

type infiniteReader struct{}

func (r *infiniteReader) Read(p []byte) (int, error) {
	p[0] = '1'
	p[1] = '2'
	p[2] = '3'
	p[3] = '4'
	return 4, nil
}

//
//

type failingWriter struct {
	*CaptureWriter
}

func (w *failingWriter) Write(p []byte) (int, error) {
	n, err := w.CaptureWriter.Write(p)
	if len(w.CaptureWriter.Body) > 5 {
		return 0, errors.New("utoh")
	}

	return n, err
}

//
//

type failingResponseWriter struct {
	numWrites int
	maxWrites int
	err       error
}

func NewFailingResponseWriter(maxWrites int, err error) *failingResponseWriter {
	return &failingResponseWriter{
		maxWrites: maxWrites,
		err:       err,
	}
}

func (r *failingResponseWriter) WriteHeader(int)     {}
func (r *failingResponseWriter) Header() http.Header { return http.Header{} }

func (r *failingResponseWriter) Write(b []byte) (int, error) {
	r.numWrites++

	if r.numWrites >= r.maxWrites {
		return 0, r.err
	}

	return len(b), nil
}

//
//

type decoratedCaptureWriter struct {
	*CaptureWriter
	closeChan chan bool
	flushCh   chan struct{}
}

func (r *decoratedCaptureWriter) CloseNotify() <-chan bool {
	return r.closeChan
}

func (r *decoratedCaptureWriter) Flush() {
	if r.flushCh != nil {
		r.flushCh <- struct{}{}
	}
}
