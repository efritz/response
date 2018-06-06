package response

import (
	"io"
)

// Serialize reads the entire response and returns it as a byte slice.
// This method is meant for testing and a subsequent call to WriteTo
// will write an empty body.
func Serialize(r Response) ([]byte, error) {
	w := NewCaptureWriter(0)

	var err error
	r.AddCallback(func(e error) { err = e })
	r.WriteTo(w)

	return w.Body, err
}

//
// Private Helpers

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
