package request

import (
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/pkg/errors"
)

type Request struct {
	Transport http.RoundTripper

	baseURL url.URL
	c       *http.Client
}

func NewRequest(baseURL string) (*Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse URL '%s'", baseURL)
	}

	r := &Request{
		baseURL: *u,
	}

	return r, nil
}

func (r *Request) ProcessGet(query string) (io.ReadCloser, bool, error) {
	body, code, err := r.process(http.MethodGet, query, nil)
	if err != nil {
		return nil, false, err
	}

	switch code {
	case http.StatusOK:
		return body, true, nil
	default:
		return body, false, nil
	}
}

func (r *Request) process(verb, query string, body io.Reader) (io.ReadCloser, int, error) {
	u := r.baseURL
	u.Path = path.Join(u.Path, query)
	completeURL := u.String()

	req, err := http.NewRequest(verb, completeURL, body)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "Failed to create request for '%s %s'", verb, completeURL)
	}

	c := http.Client{
		Transport: r.Transport,
	}

	resp, err := c.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, 0, errors.Wrapf(err, "Failed to perform request")
	}

	return resp.Body, resp.StatusCode, nil
}
