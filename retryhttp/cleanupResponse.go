package retryhttp

import (
	"io"
	"net/http"

	"github.com/xmidt-org/retry"
)

// CleanupResponse is an OnAttempt that drains and closes the response body
// between each attempt.  If the given attempt is the last one, including if
// it represents an error, the response is left as is for the Client's caller
// to deal with.
func CleanupResponse(a retry.Attempt[*http.Response]) {
	if !a.Done() && a.Result != nil && a.Result.Body != nil {
		io.Copy(io.Discard, a.Result.Body)
		a.Result.Body.Close()
		a.Result.Body = nil
	}
}
