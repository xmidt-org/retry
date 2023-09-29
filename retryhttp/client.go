// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"io"
	"net/http"
)

// Client is a function that can execute an HTTP transaction.
// http.Client.Do and http.RoundTripper.RoundTrip work for this.
type Client func(*http.Request) (*http.Response, error)

// cleanup handles draining and closing a client HTTP response.
func cleanup(response *http.Response) {
	if response != nil && response.Body != nil {
		io.Copy(io.Discard, response.Body)
		response.Body.Close()
	}
}
