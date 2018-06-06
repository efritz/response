package response

import (
	"io"
	"net/http"
)

type (
	streamConfig struct {
		progress        chan<- int
		flushAfterWrite bool
	}

	// StreamConfigFunc is a function used to configure the Stream constructor.
	StreamConfigFunc func(*streamConfig)
)

// WithProgressChan instructs Stream to send the number of bytes written
// to this chan after every successful chunk of data is written. Consumers
// of this channel should be efficient as this write will block progress.
// This channel is closed by the WriteTo method.
func WithProgressChan(progress chan<- int) StreamConfigFunc {
	return func(s *streamConfig) { s.progress = progress }
}

// WithFlush instructs Stream to call the writer's Flush method after
// every successful chunk of data is written.
func WithFlush() StreamConfigFunc {
	return func(s *streamConfig) { s.flushAfterWrite = true }
}

// Stream creates a response that writes the data from the given reader.
// The reader is closed once all data is consumed, an error is encountered,
// or the client disconnects.
func Stream(rc io.ReadCloser, configs ...StreamConfigFunc) Response {
	config := &streamConfig{
		progress:        nil,
		flushAfterWrite: false,
	}

	for _, f := range configs {
		f(config)
	}

	return NewResponse(func(w io.Writer) error {
		defer rc.Close()

		if config.progress != nil {
			defer close(config.progress)
		}

		buffer := make([]byte, 32*1024)

		for !isClosed(w) {
			n, err := moveChunk(rc, w, buffer)
			if err != nil {
				if err == io.EOF {
					break
				}

				return err
			}

			if f, ok := w.(http.Flusher); ok && config.flushAfterWrite {
				f.Flush()
			}

			if config.progress != nil {
				config.progress <- n
			}
		}

		return nil
	})
}

// isClosed returns true if the given writer is a CloseNotifier
// and the remote end has already disconnected.
func isClosed(w io.Writer) bool {
	if cn, ok := w.(http.CloseNotifier); ok {
		select {
		case <-cn.CloseNotify():
			return true
		default:
		}
	}

	return false
}

// moveChunk reads a chunk from r and writes it to w using the given
// buffer as scratch space. Returns the number of bytes read and the
// error from either read or write operations.
func moveChunk(r io.Reader, w io.Writer, buffer []byte) (int, error) {
	n, err := r.Read(buffer)
	if n > 0 {
		if err = writeAll(w, buffer[0:n]); err != nil {
			return 0, err
		}
	}

	return n, err
}
