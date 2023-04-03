package common

import "net/http"

// HTTPClient interface
type IHttpRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

