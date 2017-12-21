package response

import "io"

type captureWriter struct {
	chunkSize int
	callCount int
	buffer    []byte
}

func newCaptureWriter(chunkSize int) *captureWriter {
	return &captureWriter{
		chunkSize: chunkSize,
		callCount: 0,
		buffer:    []byte{},
	}
}

func (w *captureWriter) Write(p []byte) (int, error) {
	if w.chunkSize < len(p) && w.chunkSize > 0 {
		p = p[:w.chunkSize]
	}

	w.callCount++
	w.buffer = append(w.buffer, p...)
	return len(p), nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		data = data[n:]
	}

	return nil
}
