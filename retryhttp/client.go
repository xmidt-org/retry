package retryhttp

import (
	"context"
	"io"
	"net/http"

	"github.com/xmidt-org/retry"
)

// cleanupResponse is an OnAttempt that drains and closes the response body
// between each attempt.  If the given attempt is the last one, including if
// it represents an error, the response is left as is for the Client's caller
// to deal with.
func cleanupResponse(a retry.Attempt[*http.Response]) {
	if !a.Done() && a.Result != nil && a.Result.Body != nil {
		io.Copy(io.Discard, a.Result.Body)
		a.Result.Body.Close()
		a.Result.Body = nil
	}
}

// HTTPClient is the required behavior of anything that can execute
// HTTP transactions.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Requester is a strategy for modifying an HTTP request before it is
// attempted.  The supplied request can be modified and returned, or a new
// request may be created.  If a new request is created, it should incorporate
// the context from the supplied request.
//
// Callers can use one or more Requesters to tailor the request prior to
// each attempt.  Authorization is an example of something that might need
// to be supplied before each attempt instead of in the original request.
type Requester func(*http.Request) *http.Request

// ClientOption is a configurable option for a Client.
type ClientOption interface {
	apply(*Client) error
}

type clientOptionFunc func(*Client) error

func (cof clientOptionFunc) apply(c *Client) error { return cof(c) }

// WithHTTPClient associates the given HTTP client with the Client being created.
// If this option is not supplied, http.DefaultClient is used.
func WithHTTPClient(hc HTTPClient) ClientOption {
	return clientOptionFunc(func(c *Client) error {
		c.hc = hc
		return nil
	})
}

// WithRunner sets the retry.Runner for the Client.  An OnAttempt is appended
// that cleans up HTTP responses between attempts, leaving the last response
// untouched so that callers may inspect it.
func WithRunner(opts ...retry.RunnerOption[*http.Response]) ClientOption {
	return clientOptionFunc(func(c *Client) error {
		opts = append(opts, retry.WithOnAttempt(cleanupResponse))
		runner, err := retry.NewRunner(opts...)
		if err != nil {
			return err
		}

		c.runner = runner
		return nil
	})
}

// WithRequesters appends Requester strategies to the client.  Multiple uses of
// this option are cumulative.
func WithRequesters(r ...Requester) ClientOption {
	return clientOptionFunc(func(c *Client) error {
		c.requesters = append(c.requesters, r...)
		return nil
	})
}

type Client struct {
	hc         HTTPClient
	runner     retry.Runner[*http.Response]
	requesters []Requester
}

func NewClient(opts ...ClientOption) (c *Client, err error) {
	c = new(Client)
	for _, o := range opts {
		if err = o.apply(c); err != nil {
			break
		}
	}

	if err == nil && c.hc == nil {
		c.hc = http.DefaultClient
	}

	if err == nil && c.runner == nil {
		// since no retries will be performed, don't bother adding cleanupResponse
		c.runner, err = retry.NewRunner[*http.Response]()
	}

	if err != nil {
		c = nil
	}

	return
}

func (c *Client) newTask(original *http.Request) retry.Task[*http.Response] {
	// save off any GetBody strategy and use it to repopulate the Body
	// before each attempt
	getBody := original.GetBody

	return func(ctx context.Context) (response *http.Response, err error) {
		request := original.Clone(ctx)
		request.Body = nil // use this to detect if a Requester set a body

		for _, r := range c.requesters {
			request = r(request)
		}

		// if a Requester set a Body, leave it alone
		if request.Body == nil && getBody != nil {
			request.Body, err = getBody()
		}

		if err == nil {
			response, err = c.hc.Do(request)
		}

		return
	}
}

// Do executes the HTTP transaction specified by the request.  Any policy factory
// specified to WithRunner will determine if and how often the HTTP transaction
// is retried.
//
// The context of the original request is used as the absolute bounds for all retries.
// If the request was not created with a context, then only the retry policy will govern
// any limits.
//
// If a body should be transmitted with each attempt, a caller must either specify
// a Request.GetBody or use a Requester to set the body.  A Requester will cause the
// use of the same body for all requests sent through this client, which may not be desirable.
// Set the request's GetBody, just as one would for redirects, to allow per-request bodies.
// Note that http.NewRequest and http.NewRequestWithContext both set GetBody for common
// standard library types.
func (c *Client) Do(original *http.Request) (*http.Response, error) {
	return c.runner.Run(
		original.Context(),
		c.newTask(original),
	)
}
