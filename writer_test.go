package response

import (
	"errors"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type WriterSuite struct{}

func (s *WriterSuite) TestWriteAll(t sweet.T) {
	var (
		w    = NewCaptureWriter(2)
		data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	)

	err := writeAll(w, data)
	Expect(err).To(BeNil())
	Expect(w.Body).To(Equal(data))
	Expect(w.numWrites).To(Equal(5))
}

func (s *WriterSuite) TestWriteAllError(t sweet.T) {
	var (
		w    = &failingWriter{NewCaptureWriter(2)}
		data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	)

	Expect(writeAll(w, data)).To(MatchError("utoh"))
}

//
// Writer that returns an error

type failingWriter struct {
	*CaptureWriter
}

func (w *failingWriter) Write(p []byte) (int, error) {
	n, err := w.CaptureWriter.Write(p)
	if len(w.CaptureWriter.Body) > 5 {
		return 0, errors.New("utoh")
	}

	return n, err
}
