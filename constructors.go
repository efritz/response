package response

import (
	"encoding/json"
	"errors"
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

// ErrClientDisconnect occurs when a client disconnects during a streaming response.
var ErrClientDisconnect = errors.New("client disconnected")

// Respond creates a response with the given body.
func Respond(data []byte) *Response {
	return newResponse(func(w io.Writer) error {
		return writeAll(w, data)
	})
}

// Empty creates an empty response with the given status code.
func Empty(statusCode int) *Response {
	return Respond(nil).SetStatusCode(statusCode)
}

// JSON creates a response with the data serialized as JSON for the body.
func JSON(data interface{}) *Response {
	body, _ := json.Marshal(data)
	return Respond(body).SetHeader("Content-Type", "application/json")
}

// Stream creates a response that writes the data from the given reader.
// The reader is closed once all data is consumed, an error is encountered,
// or the client disconnects.
func Stream(rc io.ReadCloser, configs ...StreamConfigFunc) *Response {
	config := &streamConfig{
		progress:        nil,
		flushAfterWrite: false,
	}

	for _, f := range configs {
		f(config)
	}

	return newResponse(func(w io.Writer) error {
		defer rc.Close()

		if config.progress != nil {
			defer close(config.progress)
		}

		if cn, ok := w.(http.CloseNotifier); ok {
			go func() {
				<-cn.CloseNotify()
				rc.Close()
			}()
		}

		buffer := make([]byte, 32*1024)

		for {
			n, err := rc.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return nil
				}

				return err
			}

			if err := writeAll(w, buffer[0:n]); err != nil {
				return err
			}

			if f, ok := w.(http.Flusher); ok && config.flushAfterWrite {
				f.Flush()
			}

			if config.progress != nil {
				config.progress <- n
			}
		}
	})
}

// WithProgressChan instructs Stream to send the number of bytes written
// to this chan after every successful chunk of data is written. Consumers
// of this channel should be efficient as this write will block progress.
func WithProgressChan(progress chan<- int) StreamConfigFunc {
	return func(s *streamConfig) { s.progress = progress }
}

// WithFlush instructs Stream to call the writer's Flush method after
// every successful chunk of data is written.
func WithFlush() StreamConfigFunc {
	return func(s *streamConfig) { s.flushAfterWrite = true }
}
