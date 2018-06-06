package response

import (
	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type IOUtilSuite struct{}

func (s *IOUtilSuite) TestWriteAll(t sweet.T) {
	var (
		w    = NewCaptureWriter(2)
		data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	)

	err := writeAll(w, data)
	Expect(err).To(BeNil())
	Expect(w.Body).To(Equal(data))
	Expect(w.numWrites).To(Equal(5))
}

func (s *IOUtilSuite) TestWriteAllError(t sweet.T) {
	var (
		w    = &failingWriter{NewCaptureWriter(2)}
		data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	)

	Expect(writeAll(w, data)).To(MatchError("utoh"))
}
