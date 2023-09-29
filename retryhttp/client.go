// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"net/http"

	"github.com/xmidt-org/retry"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientOption interface {
	apply(*Client) error
}

type clientOptionFunc func(*Client) error

func (cof clientOptionFunc) apply(c *Client) error { return cof(c) }

func WithHTTPClient(hc HTTPClient) ClientOption {
	return clientOptionFunc(func(c *Client) error {
		c.hc = hc
		return nil
	})
}

func WithRunner(ro ...retry.RunnerOption) ClientOption {
	return clientOptionFunc(func(c *Client) (err error) {
		c.runner, err = retry.NewRunner[*http.Response](ro...)
		return
	})
}

type Client struct {
	hc     HTTPClient
	runner retry.Runner[*http.Response]
}

func NewClient(opts ...ClientOption) (c *Client, err error) {
	c = new(Client)

	for _, o := range opts {
		err = o.apply(c)
	}

	return
}

func (c *Client) httpClient() HTTPClient {
	if c.hc != nil {
		return c.hc
	}

	return http.DefaultClient
}

func (c *Client) Do(original *http.Request) (*http.Response, error) {
	return nil, nil // TODO
}
