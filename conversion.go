package response

import "net/http"

type (
	// HandlerFunc is an analog of an http.HandlerFunc that returns a
	// response object instead of writing directly to a ResponseWriter.
	HandlerFunc func(*http.Request) Response
)

// Convert converts a HandlerFunc to an http.HandlerFunc.
func Convert(f HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f(r).WriteTo(w)
	})
}
