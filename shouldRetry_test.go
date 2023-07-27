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
		suite.False(ShouldRetry(nil, nil))
	})

	suite.Run("WithPredicate", func() {
		suite.False(ShouldRetry(nil, func(error) bool { return true }))
	})
}

func (suite *ShouldRetrySuite) TestShouldRetryable() {
	suite.Run("NoPredicate", func() {
		suite.True(
			ShouldRetry(
				SetRetryable(errors.New("expected"), true),
				nil,
			),
		)

		suite.False(
			ShouldRetry(
				SetRetryable(errors.New("expected"), false),
				nil,
			),
		)
	})

	// the predicate should be ignored in favor of the retryable error
	suite.Run("WithPredicate", func() {
		suite.True(
			ShouldRetry(
				SetRetryable(errors.New("expected"), true),
				func(error) bool { return false },
			),
		)

		suite.False(
			ShouldRetry(
				SetRetryable(errors.New("expected"), false),
				func(error) bool { return true },
			),
		)
	})
}

func (suite *ShouldRetrySuite) TestPredicate() {
	expectedErr := errors.New("expected")

	suite.True(
		ShouldRetry(
			expectedErr,
			func(actualErr error) bool {
				suite.Same(expectedErr, actualErr)
				return true
			},
		),
	)

	suite.False(
		ShouldRetry(
			expectedErr,
			func(actualErr error) bool {
				suite.Same(expectedErr, actualErr)
				return false
			},
		),
	)
}

func (suite *ShouldRetrySuite) TestFallthrough() {
	suite.False(
		ShouldRetry(
			errors.New("no other way to determine retryability for this error"),
			nil,
		),
	)
}

func TestShouldRetry(t *testing.T) {
	suite.Run(t, new(ShouldRetrySuite))
}
