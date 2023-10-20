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
*/
package retryhttp
