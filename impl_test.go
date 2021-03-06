package response

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ImplementationSuite struct{}

func (s *ImplementationSuite) TestSetters(t sweet.T) {
	resp := newResponse(nil)
	Expect(resp.StatusCode()).To(Equal(http.StatusOK))
	Expect(resp.SetStatusCode(http.StatusNotFound)).To(Equal(resp))
	Expect(resp.StatusCode()).To(Equal(http.StatusNotFound))

	Expect(resp.SetHeader("X-Foo", "bar")).To(Equal(resp))
	Expect(resp.Header("X-Foo")).To(Equal("bar"))
	Expect(resp.SetHeader("X-Foo", "baz")).To(Equal(resp))
	Expect(resp.Header("X-Foo")).To(Equal("baz"))

	Expect(resp.AddHeader("X-Foo", "bonk")).To(Equal(resp))
	Expect(resp.Header("X-Foo")).To(Equal("baz"))

	w := httptest.NewRecorder()
	resp.WriteTo(w)
	Expect(w.Header()["X-Foo"]).To(Equal([]string{"baz", "bonk"}))
}

func (s *ImplementationSuite) TestDecorateWriter(t sweet.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(`abcdefg`)))
	resp := Stream(r)

	resp.DecorateWriter(func(w io.Writer) io.Writer {
		return WriterFunc(func(p []byte) (int, error) { return w.Write(upperBytes(p)) })
	})

	_, body, err := Serialize(resp)
	Expect(err).To(BeNil())
	Expect(body).To(Equal([]byte("ABCDEFG")))
}

func (s *ImplementationSuite) TestDecorateWriterCloseError(t sweet.T) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(`abcdefg`)))
	resp := Stream(r)

	resp.DecorateWriter(func(w io.Writer) io.Writer {
		write := func(p []byte) (int, error) {
			return w.Write(upperBytes(p))
		}

		return &failCloser{WriterFunc(write)}
	})

	_, _, err := Serialize(resp)
	Expect(err).To(MatchError("utoh"))
}

func upperBytes(p []byte) []byte {
	return []byte(strings.ToUpper(string(p)))
}

func (s *ImplementationSuite) TestDecorateWriterCloseNotifier(t sweet.T) {
	resp := Stream(ioutil.NopCloser(&infiniteReader{}))

	resp.DecorateWriter(func(w io.Writer) io.Writer {
		return WriterFunc(func(p []byte) (int, error) { return w.Write(p) })
	})

	var (
		ch        = make(chan string)
		closeChan = make(chan bool)
	)

	go func() {
		defer close(ch)

		w := &decoratedRecorder{
			httptest.NewRecorder(),
			closeChan,
			nil,
		}

		resp.WriteTo(w)
		ch <- string(w.Body.Bytes())
	}()

	close(closeChan)
	Eventually(ch).Should(Receive())
}

func (s *ImplementationSuite) TestMultipleWriteToCallsPanics(t sweet.T) {
	resp := JSON(nil)

	// This one is fine
	resp.WriteTo(httptest.NewRecorder())

	// This one is not
	Expect(func() { resp.WriteTo(httptest.NewRecorder()) }).To(Panic())
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

type failCloser struct {
	io.Writer
}

func (c *failCloser) Close() error {
	return fmt.Errorf("utoh")
}
