package apigatewayiface

import (
	"net/http"
	"net/url"
)

// Service is an interface for making signed requests to API Gateway.
type Service interface {
	Get(string, url.Values) (*http.Response, error)
	Put(string, interface{}) (*http.Response, error)
}
