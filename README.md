# Response

[![GoDoc](https://godoc.org/github.com/efritz/response?status.svg)](https://godoc.org/github.com/efritz/response)
[![Build Status](https://secure.travis-ci.org/efritz/response.png)](http://travis-ci.org/efritz/response)
[![Maintainability](https://api.codeclimate.com/v1/badges/69a8d691fd23fd17cc35/maintainability)](https://codeclimate.com/github/efritz/response/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/69a8d691fd23fd17cc35/test_coverage)](https://codeclimate.com/github/efritz/response/test_coverage)

Go library for wrapping HTTP responses.

## Example

The basic usage changes the `http.HandleFunc` slightly. Instead of taking a request
and a response writer, you take a request and return a response object. This allows
a more linear flow of logic through http handlers as responses are values instead of
side-effects.

The basic usage is shown below.

```go
emptyHandler := func (r *http.Request) response.Response {
    return response.Empty(http.StatusNoContent)
}

http.HandleFunc("/empty", response.Convert(emptyHandler))
```

There are several convenience constructors like the JSON response constructor shown
below. This example serializes a map into a JSON object and sets additional headers.

```go
func (r *http.Request) response.Response {
    resp := response.JSON(map[string]interface{}{
        "foo": "bar",
        "baz": []int{3, 4, 5},
    })

    resp.AddHeader("Location", "/foo/bar/baz")
    resp.AddHeader("X-Request-ID", "1234-567")
    return resp
}
```

There is also support for attaching a reader for streaming a response body. This is
useful if responses are very large or infinite (for example, a media server or an
endpoint that returns server-sent events).

```go
func (r *http.Request) response.Response {
    ch := make(chan int)

    go func() {
        for n := range ch {
            fmt.Printf("Sent an additional %d bytes of data\n", n)
        }
    }()

    return response.Stream(
        reader,               // Body content (read closer)
        WithProgressChan(ch), // Monitor how much data was sent to client
        WithFlush(),          // Enable flushing after every write (32k chunks)
    )
}
```

The `Stream` constructor will watch for client disconnect and discontinue calling
the reader for additional data.

## License

Copyright (c) 2017 Eric Fritz

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
