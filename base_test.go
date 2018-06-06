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
		data2 = []byte{2, 4, 6, 8, 0}
	)

	r1 := Respond(data1)
	Expect(r1.StatusCode()).To(Equal(http.StatusOK))
	Expect(Serialize(r1)).To(Equal(data1))

	r2 := Respond(data2)
	Expect(r2.StatusCode()).To(Equal(http.StatusOK))
	Expect(Serialize(r2)).To(Equal(data2))
}

func (s *BaseSuite) TestEmpty(t sweet.T) {
	r1 := Empty(http.StatusNotFound)
	Expect(r1.StatusCode()).To(Equal(http.StatusNotFound))
	Expect(Serialize(r1)).To(BeEmpty())

	r2 := Empty(http.StatusUnauthorized)
	Expect(r2.StatusCode()).To(Equal(http.StatusUnauthorized))
	Expect(Serialize(r2)).To(BeEmpty())
}

func (s *BaseSuite) TestJSON(t sweet.T) {
	payload := SampleJSON{
		PropertyA: "foo",
		PropertyB: "bar",
		PropertyC: "baz",
	}

	r := JSON(payload)
	Expect(r.Header("Content-Type")).To(Equal("application/json"))
	Expect(Serialize(r)).To(MatchJSON(`{"prop_a":"foo","prop_b":"bar","prop_c":"baz"}`))
}

type SampleJSON struct {
	PropertyA string `json:"prop_a"`
	PropertyB string `json:"prop_b"`
	PropertyC string `json:"prop_c"`
}
