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

type ConstructorsSuite struct{}

func (s *ConstructorsSuite) TestRespond(t sweet.T) {
	var (
		data1 = []byte{1, 3, 5, 7, 9}
		data2 = []byte{2, 4, 6, 8, 0}
	)

	r1 := Respond(data1)
	Expect(r1.StatusCode).To(Equal(http.StatusOK))
	Expect(Serialize(r1)).To(Equal(data1))

	r2 := Respond(data2)
	Expect(r2.StatusCode).To(Equal(http.StatusOK))
	Expect(Serialize(r2)).To(Equal(data2))
}

func (s *ConstructorsSuite) TestEmpty(t sweet.T) {
	r1 := Empty(http.StatusNotFound)
	Expect(r1.StatusCode).To(Equal(http.StatusNotFound))
	Expect(Serialize(r1)).To(BeEmpty())

	r2 := Empty(http.StatusUnauthorized)
	Expect(r2.StatusCode).To(Equal(http.StatusUnauthorized))
	Expect(Serialize(r2)).To(BeEmpty())
}

func (s *ConstructorsSuite) TestJSON(t sweet.T) {
	payload := SampleJSON{
		PropertyA: "foo",
		PropertyB: "bar",
		PropertyC: "baz",
	}

	r := JSON(payload)
	Expect(r.GetHeader("Content-Type")).To(Equal("application/json"))
	Expect(Serialize(r)).To(MatchJSON(`{"prop_a":"foo","prop_b":"bar","prop_c":"baz"}`))
}

func (s *ConstructorsSuite) TestStream(t sweet.T) {
	var (
		data   = makeData()
		closer = &Closer{bytes.NewReader(data), false}
		resp   = Stream(closer)
	)

	Expect(closer.closed).To(BeFalse())
	Expect(Serialize(resp)).To(Equal(data))
	Expect(closer.closed).To(BeTrue())
}

func (s *ConstructorsSuite) TestStreamDisconnect(t sweet.T) {
	var (
		data       = makeData()
		closeChan  = make(chan bool)
		progressCh = make(chan int)
		writer     = &ResponseWriter{httptest.NewRecorder(), closeChan, nil}
		resp       = Stream(
			&Closer{bytes.NewReader(data), false},
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

	body := writer.Result().Body
	defer body.Close()
	written, _ := ioutil.ReadAll(body)

	Expect(len(written)).To(Or(
		Equal(len(data)/8*3),
		Equal(len(data)/8*4),
	))

	Expect(written).To(Equal(data[:len(written)]))
}

func (s *ConstructorsSuite) TestStreamWriteError(t sweet.T) {
	var (
		errors      = make(chan error, 1)
		expectedErr = fmt.Errorf("utoh")
		handler     = func(err error) { errors <- err }
		writer      = NewFailingResponseWriter(3, expectedErr)
		resp        = Stream(&Closer{bytes.NewReader(makeData()), false})
	)

	defer close(errors)

	resp.AddCallback(handler)
	resp.WriteTo(writer)

	Eventually(errors).Should(Receive(Equal(expectedErr)))
	Consistently(errors).ShouldNot(Receive())
}

func (s *ConstructorsSuite) TestStreamFlush(t sweet.T) {
	var (
		data      = makeData()
		closeChan = make(chan bool)
		flushCh   = make(chan struct{})
		writer    = &ResponseWriter{httptest.NewRecorder(), closeChan, flushCh}
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

func (s *ConstructorsSuite) TestStreamProgress(t sweet.T) {
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
// Helpers

func makeData() []byte {
	data := []byte{}
	for i := 0; i < 32*1024; i++ {
		data = append(data, []byte("12345678")...)
	}

	return data
}

//
// Serialization Data

type SampleJSON struct {
	PropertyA string `json:"prop_a"`
	PropertyB string `json:"prop_b"`
	PropertyC string `json:"prop_c"`
}

//
// Mock ReadCloser

type Closer struct {
	io.Reader
	closed bool
}

func (c *Closer) Read(buffer []byte) (int, error) {
	if c.closed {
		return 0, io.EOF
	}

	return c.Reader.Read(buffer)
}

func (c *Closer) Close() error {
	c.closed = true
	return nil
}

//
// Mock ResponseWriter with close/flush handles

type ResponseWriter struct {
	*httptest.ResponseRecorder
	closeChan chan bool
	flushCh   chan struct{}
}

func (r *ResponseWriter) CloseNotify() <-chan bool {
	return r.closeChan
}

func (r *ResponseWriter) Flush() {
	if r.flushCh != nil {
		r.flushCh <- struct{}{}
	}
}

//
// Mock ResponseWriter that can fail

type FailingResponseWriter struct {
	numWrites int
	maxWrites int
	err       error
}

func NewFailingResponseWriter(maxWrites int, err error) *FailingResponseWriter {
	return &FailingResponseWriter{
		maxWrites: maxWrites,
		err:       err,
	}
}

func (r *FailingResponseWriter) WriteHeader(int)     {}
func (r *FailingResponseWriter) Header() http.Header { return http.Header{} }

func (r *FailingResponseWriter) Write(b []byte) (int, error) {
	r.numWrites++

	if r.numWrites >= r.maxWrites {
		return 0, r.err
	}

	return len(b), nil
}
