// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ShouldRetrySuite struct {
	suite.Suite
}

func (suite *ShouldRetrySuite) TestSetRetryable() {
	for _, retryable := range []bool{true, false} {
		suite.Run(fmt.Sprintf("retryable=%v", retryable), func() {
			var (
				expectedErr = errors.New("expected")
				wrapperErr  = SetRetryable(expectedErr, retryable)
			)

			var sr ShouldRetryable
			suite.Require().ErrorAs(wrapperErr, &sr)
			suite.Equal(retryable, sr.ShouldRetry())

			suite.ErrorIs(wrapperErr, expectedErr)
		})
	}
}

func (suite *ShouldRetrySuite) TestDefaultTestErrorForRetry() {
	suite.Run("NilError", func() {
		suite.False(DefaultTestErrorForRetry(nil))
	})

	suite.Run("ShouldRetryable", func() {
		err := errors.New("expected")
		suite.True(
			DefaultTestErrorForRetry(
				SetRetryable(err, true),
			),
		)

		suite.False(
			DefaultTestErrorForRetry(
				SetRetryable(err, false),
			),
		)
	})

	suite.Run("Temporary", func() {
		suite.True(
			DefaultTestErrorForRetry(
				&net.DNSError{
					IsTemporary: true,
				},
			),
		)

		suite.False(
			DefaultTestErrorForRetry(
				&net.DNSError{
					IsTemporary: false,
				},
			),
		)
	})

	suite.Run("NonNilError", func() {
		suite.True(
			DefaultTestErrorForRetry(
				errors.New("expected"),
			),
		)
	})
}

func TestShouldRetry(t *testing.T) {
	suite.Run(t, new(ShouldRetrySuite))
}
