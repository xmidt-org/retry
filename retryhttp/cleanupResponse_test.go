// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/retry"
)

type body struct {
	*bytes.Buffer
	closed bool
}

func (b *body) Close() error {
	b.closed = true
	return nil
}

type CleanupResponseSuite struct {
	suite.Suite
}

func (suite *CleanupResponseSuite) TestDone() {
	var (
		body = &body{
			Buffer: bytes.NewBufferString("test"),
		}
	)

	CleanupResponse(retry.Attempt[*http.Response]{
		Context: context.Background(),
		Next:    0, // indicates no further attempts
		Result: &http.Response{
			Body: body,
		},
	})

	suite.Equal("test", body.Buffer.String())
	suite.False(body.closed)
}

func (suite *CleanupResponseSuite) TestNotDone() {
	var (
		body = &body{
			Buffer: bytes.NewBufferString("test"),
		}
	)

	CleanupResponse(retry.Attempt[*http.Response]{
		Context: context.Background(),
		Next:    15 * time.Second,
		Result: &http.Response{
			Body: body,
		},
	})

	suite.Empty(body.Buffer.String())
	suite.True(body.closed)
}

func TestCleanupResponse(t *testing.T) {
	suite.Run(t, new(CleanupResponseSuite))
}
