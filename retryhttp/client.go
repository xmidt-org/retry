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

// WithPolicyFactory specifies a PolicyFactory specifically for an HTTP
// client Runner.
func WithPolicyFactory(pf retry.PolicyFactory) retry.RunnerOption[*http.Response] {
	return retry.WithPolicyFactory[*http.Response](pf)
}

// WithShouldRetry specifies a should retry predicate specifically for an
// HTTP client Runner.
func WithShouldRetry(sr retry.ShouldRetry[*http.Response]) retry.RunnerOption[*http.Response] {
	return retry.WithShouldRetry(sr)
}

// WithOnAttempt specifies an OnAttempt callback specifically for an
// HTTP client runner.
func WithOnAttempt(f retry.OnAttempt[*http.Response]) retry.RunnerOption[*http.Response] {
	return retry.WithOnAttempt(f)
}

// HTTPClient is the required behavior of the clientside of HTTP transactions.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// ClientOption is a configurable option for creating a Client.
type ClientOption interface {
	apply(*Client) error
}

type clientOptionFunc func(*Client) error

func (cof clientOptionFunc) apply(c *Client) error { return cof(c) }

// WithHTTPClient specifies the HTTPClient used to execute HTTP transactions.
func WithHTTPClient(hc HTTPClient) ClientOption {
	return clientOptionFunc(func(c *Client) error {
		c.hc = hc
		return nil
	})
}

// Requester is a strategy for modifying an HTTP request.  The original request
// may be returned with some modifications, or an entirely new request object may
// be returned instead.  If a new request object is returned, implementations must
// take care to use the context of the original request.
//
// Clients allow Requesters to modify each retry's request prior to execution.
type Requester func(*http.Request) *http.Request

// WithRequesters associates Requester strategies with a Client.  The Requesters
// are appended to any that were previously specified.  Requesters are executed
// in the order presented by this option.
func WithRequesters(rfs ...Requester) ClientOption {
	return clientOptionFunc(func(c *Client) error {
		c.requesters = append(c.requesters, rfs...)
		return nil
	})
}

// WithRunner creates a Runner to use with the Client.  If this option is
// not used, the returned Client will not retry failed HTTP transactions.
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

// Client is an HTTP client that retries transactions according to a policy.
type Client struct {
	hc         HTTPClient
	requesters []Requester
	runner     retry.Runner[*http.Response]
}

// NewClient creates a new Client using the supplied options.
//
// If no WithHTTPClient option is specified, the returned Client will use
// the http.DefaultClient.
//
// If no WithRunner option is specified, the returned Client will never
// retry any HTTP transaction.
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

// transactor returns a Task that will attempt the given request.  The request is cloned,
// and any Requesters are run against that clone, prior to the attempt.
func (c *Client) transactor(original *http.Request) retry.Task[*http.Response] {
	return func(taskCtx context.Context) (*http.Response, error) {
		request := original.Clone(taskCtx)
		for _, r := range c.requesters {
			request = r(request)
		}

		return c.hc.Do(request)
	}
}

// Do executes the given request, retrying it according to the configured policy.
// The request passed to this method is used as a prototype, and is cloned for
// each attempt.  After cloning, any Requesters that were specified are allowed
// to modify the clone request, e.g. adding headers or URL parameters.
//
// After each unsuccessful attempt, the HTTP response is properly drained and closed
// in accordance with the contract of http.Client.Do.  After a successful attempt,
// the response is returned as is, without closing the response body.
func (c *Client) Do(original *http.Request) (*http.Response, error) {
	return c.runner.Run(
		original.Context(),
		c.transactor(original),
	)
}
