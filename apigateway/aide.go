package apigateway

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/cleardataeng/aidews"
)

// HTTPClient is an interface for the http.Client.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Service for making signed requests.
type Service struct {
	signer  *v4.Signer
	http    HTTPClient
	host    *url.URL
	region  *string
	headers map[string]string
}

// New returns an API with which you can make API Gateway signed requests.
func New(host *url.URL, region string, roleARN *string) *Service {
	return NewWithHeaders(host, region, roleARN, nil)
}

// NewWithHeaders returns an API with which you can make API Gateway signed requests with headers.
func NewWithHeaders(host *url.URL, region string, roleARN *string, headers map[string]string) *Service {
	s := aidews.Session(&region, roleARN)
// NewWithSession returns an API like New but with a given Session.
func NewWithSession(host *url.URL, session *session.Session) *Service {
	return &Service{
		signer: v4.NewSigner(session.Config.Credentials),
		host:   host,
		region: session.Config.Region,
		http: &http.Client{
			Timeout: time.Second * 60,
		},
	}
}

// Do signs then executes do on passed in request.
func (svc *Service) Do(req *http.Request) (*http.Response, error) {
	var body io.ReadSeeker
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	if svc.headers != nil {
		for key, value := range svc.headers {
			req.Header.Set(key, value)
		}
	}
	if _, err := svc.signer.Sign(req, body, "execute-api", *svc.region, time.Now()); err != nil {
		return nil, err
	}
	return svc.http.Do(req)
}

// Get from given path.
func (svc *Service) Get(path string, qs url.Values) (*http.Response, error) {
	u, err := svc.URL(path, qs)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	return svc.Do(req)
}

// Post to given path.
func (svc *Service) Post(path string, body interface{}) (*http.Response, error) {
	b, _ := json.Marshal(body)
	seeker := bytes.NewReader(b)
	u, err := svc.URL(path, nil)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", u.String(), seeker)
	return svc.Do(req)
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
	return svc.Do(req)
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
