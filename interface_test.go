package response

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type InterfaceSuite struct{}

func (s *InterfaceSuite) TestRoundTripSerialize(t sweet.T) {
	resp1 := JSON(map[string]interface{}{"foo": []int{1, 2, 3}})
	resp1.SetStatusCode(http.StatusAccepted)
	resp1.AddHeader("foo", "bar")
	resp1.AddHeader("foo", "baz")
	resp1.AddHeader("bar", "bonk")

	headers1, body1, err := Serialize(resp1)
	Expect(err).To(BeNil())

	resp2 := Reconstruct(resp1.StatusCode(), headers1, body1)
	headers2, body2, err := Serialize(resp2)
	Expect(err).To(BeNil())

	Expect(headers1).To(Equal(headers2))
	Expect(body1).To(Equal(body2))
	Expect(resp1.StatusCode()).To(Equal(resp2.StatusCode()))
}

func (s *InterfaceSuite) TestConvert(t sweet.T) {
	var (
		errors = make(chan error, 2)
		c1     = func(err error) { errors <- err }
		c2     = func(err error) { errors <- err }
	)

	server := httptest.NewServer(Convert(func(r *http.Request) Response {
		defer r.Body.Close()
		data, _ := ioutil.ReadAll(r.Body)

		resp := JSON(map[string]interface{}{"input": string(data)})
		resp.SetStatusCode(http.StatusAccepted)
		resp.AddHeader("X-Context", "test")
		resp.AddCallback(c1)
		resp.AddCallback(c2)
		return resp
	}))

	defer close(errors)
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, bytes.NewReader([]byte("content")))
	resp, err := http.DefaultClient.Do(req)
	Expect(err).To(BeNil())

	Eventually(errors).Should(Receive(nil))
	Eventually(errors).Should(Receive(nil))
	Consistently(errors).ShouldNot(Receive())

	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
	Expect(resp.Header.Get("X-Context")).To(Equal("test"))
	Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))
	Expect(data).To(MatchJSON(`{"input": "content"}`))
}
