package response

import (
	"net/http"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ResponseSuite struct{}

func (s *ResponseSuite) TestConvert(t sweet.T) {
	// TODO
	Expect(true).To(BeTrue())
}

func (s *ResponseSuite) TestSetters(t sweet.T) {
	resp := newResponse(nil)
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp.SetStatusCode(http.StatusNotFound)).To(Equal(resp))
	Expect(resp.StatusCode).To(Equal(http.StatusNotFound))

	Expect(resp.SetHeader("X-Foo", "bar")).To(Equal(resp))
	Expect(resp.GetHeader("X-Foo")).To(Equal("bar"))
	Expect(resp.SetHeader("X-Foo", "baz")).To(Equal(resp))
	Expect(resp.GetHeader("X-Foo")).To(Equal("baz"))

	Expect(resp.AddHeader("X-Foo", "bonk")).To(Equal(resp))
	Expect(resp.GetHeader("X-Foo")).To(Equal("baz"))
	Expect(resp.Header["X-Foo"]).To(Equal([]string{"baz", "bonk"}))
}

func (s *ResponseSuite) TestWriteTo(t sweet.T) {
	// TODO
	Expect(true).To(BeTrue())
}
