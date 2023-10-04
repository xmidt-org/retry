// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"errors"
	"fmt"
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

func (suite *ShouldRetrySuite) TestNilError() {
	suite.Run("NoPredicate", func() {
		suite.False(CheckRetry(123, nil, nil))
	})

	suite.Run("WithPredicate", func() {
		suite.False(CheckRetry(123, nil, func(int, error) bool { return true }))
	})
}

func (suite *ShouldRetrySuite) TestShouldRetryable() {
	suite.Run("NoPredicate", func() {
		suite.True(
			CheckRetry(
				123,
				SetRetryable(errors.New("expected"), true),
				nil,
			),
		)

		suite.False(
			CheckRetry(
				123,
				SetRetryable(errors.New("expected"), false),
				nil,
			),
		)
	})

	// the predicate should be used instead of the error
	suite.Run("WithPredicate", func() {
		suite.False(
			CheckRetry(
				123,
				SetRetryable(errors.New("expected"), true),
				func(int, error) bool { return false },
			),
		)

		suite.True(
			CheckRetry(
				123,
				SetRetryable(errors.New("expected"), false),
				func(int, error) bool { return true },
			),
		)
	})
}

func (suite *ShouldRetrySuite) TestPredicate() {
	expectedErr := errors.New("expected")

	suite.True(
		CheckRetry(
			123,
			expectedErr,
			func(_ int, actualErr error) bool {
				suite.Same(expectedErr, actualErr)
				return true
			},
		),
	)

	suite.False(
		CheckRetry(
			123,
			expectedErr,
			func(_ int, actualErr error) bool {
				suite.Same(expectedErr, actualErr)
				return false
			},
		),
	)
}

func (suite *ShouldRetrySuite) TestFallthrough() {
	suite.True(
		CheckRetry(
			123,
			errors.New("by default, all errors are retryable"),
			nil,
		),
	)
}

func TestShouldRetry(t *testing.T) {
	suite.Run(t, new(ShouldRetrySuite))
}
