package response

import (
	"net/http/httptest"
)

type decoratedRecorder struct {
	*httptest.ResponseRecorder
	closeChan chan bool
	flushCh   chan struct{}
}

func (r *decoratedRecorder) CloseNotify() <-chan bool {
	return r.closeChan
}

func (r *decoratedRecorder) Flush() {
	if r.flushCh != nil {
		r.flushCh <- struct{}{}
	}
}
