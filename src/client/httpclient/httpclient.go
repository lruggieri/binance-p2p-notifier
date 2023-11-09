package httpclient

import "net/http"

// Client define abstraction of http client. Useful for testing.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}
