package retryhttp

import (
	"errors"
	"net/http"

	"github.com/xmidt-org/retry"
)

// NewShouldRetry creates a retry predicate that retries any of the given
// status codes from a valid response.
//
// The returned predicate will retry any errors that provide a 'Temporary() bool'
// method that returns true.  Any other kind of error is considered fatal and
// will halt retries.
func NewShouldRetry(statusCodes ...int) retry.ShouldRetry[*http.Response] {
	codes := make(map[int]bool, len(statusCodes))
	for _, sc := range statusCodes {
		codes[sc] = true
	}

	return func(response *http.Response, err error) bool {
		type temporary interface {
			Temporary() bool
		}

		var t temporary

		switch {
		// guard against client middleware that sometimes incorrectly returns nil responses and nil errors
		case err == nil && response != nil:
			return codes[response.StatusCode]

		case errors.As(err, &t):
			return t.Temporary()

		default:
			return false
		}
	}
}
