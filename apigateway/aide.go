package apigateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/cleardataeng/aidews"
)

// HTTPClient is an interface for the http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Service for making signed requests.
type Service struct {
	signer *v4.Signer
	http   HTTPClient
	host   *url.URL
	region *string
}

// New returns an API with which you can make API Gateway signed requests.
func New(host *url.URL, region string, roleARN *string) *Service {
	s := aidews.Session(&region, roleARN)
	return &Service{
		signer: v4.NewSigner(s.Config.Credentials),
		host:   host,
		region: s.Config.Region,
		http: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// Get from given path.
func (svc *Service) Get(path string, qs url.Values) (*http.Response, error) {
	u, err := svc.URL(path, qs)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	if _, err := svc.signer.Sign(req, nil, "execute-api", *svc.region, time.Now()); err != nil {
		return nil, err
	}
	return svc.http.Do(req)
}

// Put to given path.
func (svc *Service) Put(path string, body interface{}) (*http.Response, error) {
	b, _ := json.Marshal(body)
	seeker := bytes.NewReader(b)
	u, err := svc.URL(path, nil)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("PUT", u.String(), seeker)
	if _, err := svc.signer.Sign(req, seeker, "execute-api", *svc.region, time.Now()); err != nil {
		return nil, err
	}
	return svc.http.Do(req)
}

// URL adds a valid path to the Gateway host and adds an encoded query string.
func (svc *Service) URL(path string, qs url.Values) (*url.URL, error) {
	p, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	u := svc.host.ResolveReference(p)
	u.RawQuery = qs.Encode()
	return u, nil
}
