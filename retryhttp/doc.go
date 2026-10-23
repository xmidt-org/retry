// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

/*
Package retryhttp provides a simple client that retries HTTP transactions
according to a policy.

Typical usage:

	  client, _ := retryhttp.NewClient(
		  retryhttp.NewClient( // can omit to use http.DefaultClient
			  new(http.Client),
		  ),
		  retryhttp.WithRunner(
			  retry.WithPolicyFactory(
				  retry.Config{
					  Interval: 5 * time.Second,
					  MaxRetries: 5,
				  },
			  ),
		  ),
		  retryhttp.WithShouldRetry(
			  http.StatusTooManyRequests,
		  ),
	  )

	  request, _ := http.NewRequest("GET", "https://foobar.com/", nil)
	  response, err := client.Do(request)
	  // normal response and error handling ...

# Timeouts

Each attempt runs under the request's own context, not one derived by the
runner, so the response returned by Do() stays readable after Do() returns.
Two limits apply, and they bound different things:

  - retry.Config.MaxElapsedTime decides whether another attempt starts.
    No attempt begins once it has elapsed.  It does not interrupt an
    attempt already underway.

  - http.Client.Timeout bounds each attempt, from connecting through
    reading the response body.

The worst case for one call to Do() is therefore MaxElapsedTime plus
http.Client.Timeout.  An http.Client with no Timeout leaves each attempt
bounded only by the request's context, so set one when using a runner.
*/
package retryhttp
