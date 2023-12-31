// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RunnerSuite struct {
	CommonSuite
}

func (suite *RunnerSuite) testRunNoRetries() {
	var (
		testCtx, _ = suite.testCtx()
		task       = new(mockTask[int])

		onAttempt = new(mockOnAttempt[int])
		runner    = suite.newRunner(
			WithOnAttempt[int](onAttempt.OnAttempt),
		)
	)

	task.ExpectMatch(suite.assertTestCtx, 123, nil).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result: 123,
		}), // since no retries, the non-context fields will be zeroes
	).Once()

	result, err := runner.Run(testCtx, task.Do)
	suite.Equal(123, result)
	suite.NoError(err)

	onAttempt.AssertExpectations(suite.T())
	task.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) testRunWithRetriesUntilSuccess() {
	var (
		testCtx, _ = suite.testCtx()
		task       = new(mockTask[int])

		timer     = new(mockTimer)
		onAttempt = new(mockOnAttempt[int])

		retryErr = errors.New("should retry this")
		runner   = suite.newRunner(
			WithTimer[int](timer.Timer),
			WithShouldRetry(func(_ int, err error) bool {
				return errors.Is(err, retryErr)
			}),
			WithOnAttempt[int](onAttempt.OnAttempt),
			WithPolicyFactory[int](Config{
				Interval: 5 * time.Second,
			}),
		)
	)

	timer.ExpectConstant(5*time.Second, 3).Times(3)

	task.ExpectMatch(suite.assertTestCtx, -1, retryErr).Times(3)
	task.ExpectMatch(suite.assertTestCtx, 123, nil).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  -1,
			Err:     retryErr,
			Retries: 0,
			Next:    5 * time.Second,
		}),
	).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  -1,
			Err:     retryErr,
			Retries: 1,
			Next:    5 * time.Second,
		}),
	).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  -1,
			Err:     retryErr,
			Retries: 2,
			Next:    5 * time.Second,
		}),
	).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  123,
			Retries: 3,
		}),
	).Once()

	result, err := runner.Run(testCtx, task.Do)
	suite.Equal(123, result)
	suite.NoError(err)

	timer.AssertExpectations(suite.T())
	onAttempt.AssertExpectations(suite.T())
	task.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) testRunWithRetriesAndCanceled() {
	var (
		testCtx, testCancel = suite.testCtx()
		task                = new(mockTask[int])

		timer     = new(mockTimer)
		onAttempt = new(mockOnAttempt[int])

		retryErr = errors.New("should retry this")
		runner   = suite.newRunner(
			WithTimer[int](timer.Timer),
			WithShouldRetry(func(_ int, err error) bool {
				return errors.Is(err, retryErr)
			}),
			WithOnAttempt[int](onAttempt.OnAttempt),
			WithPolicyFactory[int](Config{
				Interval: 5 * time.Second,
			}),
		)
	)

	timer.ExpectConstant(5*time.Second, 2).Times(2)
	timer.ExpectTimer(5*time.Second, nil).Once().Run(func(mock.Arguments) {
		testCancel()
	})
	timer.ExpectStop(true).Once()

	task.ExpectMatch(suite.assertTestCtx, -1, retryErr).Times(3)
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  -1,
			Err:     retryErr,
			Retries: 0,
			Next:    5 * time.Second,
		}),
	).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  -1,
			Err:     retryErr,
			Retries: 1,
			Next:    5 * time.Second,
		}),
	).Once()
	onAttempt.ExpectMatch(
		suite.newTestAttemptMatcher(Attempt[int]{
			Result:  -1,
			Err:     retryErr,
			Retries: 2,
			Next:    5 * time.Second,
		}),
	).Once()

	result, err := runner.Run(testCtx, task.Do)
	suite.Equal(0, result)
	suite.Same(testCtx.Err(), err)

	timer.AssertExpectations(suite.T())
	onAttempt.AssertExpectations(suite.T())
	task.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) TestRun() {
	suite.Run("NoRetries", suite.testRunNoRetries)
	suite.Run("WithRetriesUntilSuccess", suite.testRunWithRetriesUntilSuccess)
	suite.Run("WithRetriesAndCanceled", suite.testRunWithRetriesAndCanceled)
}

func (suite *RunnerSuite) TestOptionError() {
	var (
		expectedErr       = errors.New("expected")
		runner, actualErr = NewRunner[int](
			runnerOptionFunc[int](func(r *runner[int]) error {
				suite.NotNil(r)
				return expectedErr
			}),
		)
	)

	suite.Nil(runner)
	suite.Same(expectedErr, actualErr)
}

// WithImmediateTimer ensures that the immediate timer works properly
// when set via this option.
func (suite *RunnerSuite) TestWithImmediateTimer() {
	r := suite.newRunner(
		WithImmediateTimer[int](),
	)

	t := r.(*runner[int]).timer
	suite.Require().NotNil(t)

	ch, stop := t(time.Minute)
	select {
	case <-ch:
		// passing
	case <-time.After(time.Second):
		suite.Fail("The returned channel was not immediately signaled")
	}

	stop()
	stop() // idempotent
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}
