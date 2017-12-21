package response

import (
	"errors"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type IOSuite struct{}

func (s *IOSuite) TestWriteAll(t sweet.T) {
	var (
		w    = newCaptureWriter(2)
		data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	)

	err := writeAll(w, data)
	Expect(err).To(BeNil())
	Expect(w.buffer).To(Equal(data))
	Expect(w.callCount).To(Equal(5))
}

func (s *IOSuite) TestWriteAllError(t sweet.T) {
	var (
		w    = &failingWriter{newCaptureWriter(2)}
		data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	)

	Expect(writeAll(w, data)).To(MatchError("utoh"))
}

//
// Writer that returns an error

type failingWriter struct {
	*captureWriter
}

func (w *failingWriter) Write(p []byte) (int, error) {
	n, err := w.captureWriter.Write(p)
	if len(w.captureWriter.buffer) > 5 {
		return 0, errors.New("utoh")
	}

	return n, err
}
