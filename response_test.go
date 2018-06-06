package response

import (
	"net/http"
	"net/http/httptest"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ResponseSuite struct{}

func (s *ResponseSuite) TestSetters(t sweet.T) {
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

	w := NewCaptureWriter(0)
	resp.WriteTo(w)
	Expect(w.Header()["X-Foo"]).To(Equal([]string{"baz", "bonk"}))
}

func (s *ResponseSuite) TestMultipleWriteToCallsPanics(t sweet.T) {
	resp := JSON(nil)

	// This one is fine
	resp.WriteTo(httptest.NewRecorder())

	// This one is not
	Expect(func() { resp.WriteTo(httptest.NewRecorder()) }).To(Panic())
}
