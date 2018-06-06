package response

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type StreamSuite struct{}

func (s *StreamSuite) TestStream(t sweet.T) {
	var (
		data   = makeData()
		closer = &closer{bytes.NewReader(data), false}
		resp   = Stream(closer)
	)

	Expect(closer.closed).To(BeFalse())
	Expect(Serialize(resp)).To(Equal(data))
	Expect(closer.closed).To(BeTrue())
}

func (s *StreamSuite) TestStreamDisconnect(t sweet.T) {
	var (
		data       = makeData()
		closeChan  = make(chan bool)
		progressCh = make(chan int)
		writer     = &decoratedRecorder{httptest.NewRecorder(), closeChan, nil}
		resp       = Stream(
			&closer{bytes.NewReader(data), false},
			WithProgressChan(progressCh),
		)
	)

	go func() {
		<-progressCh
		<-progressCh
		<-progressCh
		close(closeChan)

		for range progressCh {
		}
	}()

	resp.WriteTo(writer)
	body := writer.ResponseRecorder.Body.Bytes()

	Expect(len(body)).To(Or(
		Equal(len(data)/8*3),
		Equal(len(data)/8*4),
	))

	Expect(body).To(Equal(data[:len(body)]))
}

func (s *StreamSuite) TestStreamWriteError(t sweet.T) {
	var (
		errors      = make(chan error, 1)
		expectedErr = fmt.Errorf("utoh")
		handler     = func(err error) { errors <- err }
		writer      = NewFailingResponseWriter(3, expectedErr)
		resp        = Stream(&closer{bytes.NewReader(makeData()), false})
	)

	defer close(errors)

	resp.AddCallback(handler)
	resp.WriteTo(writer)

	Eventually(errors).Should(Receive(Equal(expectedErr)))
	Consistently(errors).ShouldNot(Receive())
}

func (s *StreamSuite) TestStreamFlush(t sweet.T) {
	var (
		data      = makeData()
		closeChan = make(chan bool)
		flushCh   = make(chan struct{})
		writer    = &decoratedRecorder{httptest.NewRecorder(), closeChan, flushCh}
		resp      = Stream(
			ioutil.NopCloser(bytes.NewReader(data)),
			WithFlush(),
		)
	)

	defer close(closeChan)

	go func() {
		resp.WriteTo(writer)
		close(flushCh)
	}()

	for i := 0; i < 8; i++ {
		// 8 write calls and 8 flush calls
		Eventually(flushCh).Should(Receive())
	}

	// No more data, then closed
	Consistently(flushCh).ShouldNot(Receive())
	Eventually(flushCh).Should(BeClosed())
}

func (s *StreamSuite) TestStreamProgress(t sweet.T) {
	var (
		data       = makeData()
		progressCh = make(chan int, 10)
		resp       = Stream(
			ioutil.NopCloser(bytes.NewReader(data)),
			WithProgressChan(progressCh),
		)
	)

	Expect(Serialize(resp)).To(Equal(data))

	// Note: buffered channel should have all data, so we're
	// not using Eventually/Should method pairs - don't wait
	// or skip elements from the channel.

	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(Receive(Equal(32 * 1024)))
	Expect(progressCh).To(BeClosed())
}

//
//

func makeData() []byte {
	data := []byte{}
	for i := 0; i < 32*1024; i++ {
		data = append(data, []byte("12345678")...)
	}

	return data
}

//
//

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

func (r *failingResponseWriter) WriteHeader(int) {
}

func (r *failingResponseWriter) Header() http.Header {
	return http.Header{}
}

func (r *failingResponseWriter) Write(b []byte) (int, error) {
	r.numWrites++

	if r.numWrites >= r.maxWrites {
		return 0, r.err
	}

	return len(b), nil
}
