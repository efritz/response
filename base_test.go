package response

import (
	"net/http"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type BaseSuite struct{}

func (s *BaseSuite) TestRespond(t sweet.T) {
	var (
		data1 = []byte{1, 3, 5, 7, 9}
		data2 = []byte{2, 4, 6, 8, 0, 12}
	)

	r1 := Respond(data1)
	Expect(r1.StatusCode()).To(Equal(http.StatusOK))
	headers, body, err := Serialize(r1)
	Expect(err).To(BeNil())
	Expect(body).To(Equal(data1))
	Expect(headers.Get("Content-Length")).To(Equal("5"))

	r2 := Respond(data2)
	Expect(r2.StatusCode()).To(Equal(http.StatusOK))
	headers, body, err = Serialize(r2)
	Expect(err).To(BeNil())
	Expect(body).To(Equal(data2))
	Expect(headers.Get("Content-Length")).To(Equal("6"))
}

func (s *BaseSuite) TestEmpty(t sweet.T) {
	r1 := Empty(http.StatusNotFound)
	Expect(r1.StatusCode()).To(Equal(http.StatusNotFound))
	headers, body, _ := Serialize(r1)
	Expect(body).To(BeEmpty())
	Expect(headers.Get("Content-Length")).To(Equal("0"))

	r2 := Empty(http.StatusUnauthorized)
	Expect(r2.StatusCode()).To(Equal(http.StatusUnauthorized))
	headers, body, _ = Serialize(r2)
	Expect(body).To(BeEmpty())
	Expect(headers.Get("Content-Length")).To(Equal("0"))
}

func (s *BaseSuite) TestJSON(t sweet.T) {
	payload := SampleJSON{
		PropertyA: "foo",
		PropertyB: "bar",
		PropertyC: "baz",
	}

	r := JSON(payload)
	Expect(r.Header("Content-Type")).To(Equal("application/json"))
	headers, body, err := Serialize(r)
	Expect(err).To(BeNil())
	Expect(body).To(MatchJSON(`{"prop_a":"foo","prop_b":"bar","prop_c":"baz"}`))
	Expect(headers.Get("Content-Length")).To(Equal("46"))
}

type SampleJSON struct {
	PropertyA string `json:"prop_a"`
	PropertyB string `json:"prop_b"`
	PropertyC string `json:"prop_c"`
}
