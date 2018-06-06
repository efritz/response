package response

import "io"

type (
	// WriterFunc is a function
	WriterFunc func([]byte) (int, error)

	// closeableWriter bundles CloseNotify with an io.Writer.
	closeableWriter struct {
		io.Writer
		ch chan bool
	}
)

// CloseNotify returns a channel that closes when the remote end
// disconnects.
func (w *closeableWriter) CloseNotify() <-chan bool {
	return w.ch
}

// Write implements the io.Writer interface.
func (f WriterFunc) Write(p []byte) (int, error) {
	return f(p)
}

// writeAll writes all content in the buffer to the given writer.
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
