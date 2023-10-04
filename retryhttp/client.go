// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"context"
	"io"
	"net/http"

	"github.com/xmidt-org/retry"
)

// cleanupResponse is a retry.OnAttempt callback that cleans up HTTP responses
// for failed attempts.  If the attempt represents the last time the HTTP request
// is tried, even if it is in error, this function does nothing.
func cleanupResponse(a retry.Attempt[*http.Response]) {
	if !a.Done() && a.Result != nil && a.Result.Body != nil {
		io.Copy(io.Discard, a.Result.Body)
		a.Result.Body.Close()
	}
}

func WithPolicyFactory(pf retry.PolicyFactory) retry.RunnerOption[*http.Response] {
	return retry.WithPolicyFactory[*http.Response](pf)
}

func WithShouldRetry(sr retry.ShouldRetry[*http.Response]) retry.RunnerOption[*http.Response] {
	return retry.WithShouldRetry(sr)
}

func WithOnAttempt(f retry.OnAttempt[*http.Response]) retry.RunnerOption[*http.Response] {
	return retry.WithOnAttempt(f)
}

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

func WithRunner(ro ...retry.RunnerOption[*http.Response]) ClientOption {
	return clientOptionFunc(func(c *Client) (err error) {
		ro = append(
			ro,
			WithOnAttempt(cleanupResponse), // ensure that the response is cleaned up after failed attempts
		)

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

	if c.hc == nil {
		c.hc = http.DefaultClient
	}

	if c.runner == nil {
		c.runner, err = retry.NewRunner[*http.Response](
			WithOnAttempt(cleanupResponse),
		)
	}

	return
}

func (c *Client) transactor(original *http.Request) retry.Task[*http.Response] {
	return func(taskCtx context.Context) (*http.Response, error) {
		request := original.Clone(taskCtx)
		return c.hc.Do(request)
	}
}

func (c *Client) Do(original *http.Request) (*http.Response, error) {
	return c.runner.Run(
		original.Context(),
		c.transactor(original),
	)
}
