package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type taskSuite struct {
	suite.Suite
}

func (suite *taskSuite) testContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

type RunnerSuite struct {
	taskSuite
}

func (suite *RunnerSuite) newRunner(o ...RunnerOption) Runner {
	r, err := NewRunner(o...)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)
	return r
}

func (suite *RunnerSuite) doRunFirstAttemptSuccess(f func(r Runner)) {
	sleep := new(mockSleep)
	r := suite.newRunner(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
	)

	r.(*runner).sleep = sleep.Sleep
	f(r)

	sleep.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) doRunNoPolicyFactory(f func(error, Runner)) {
	expectedErr := errors.New("expected")
	sleep := new(mockSleep)
	r := suite.newRunner()

	// no policy factory means no retries

	r.(*runner).sleep = sleep.Sleep
	f(expectedErr, r)

	sleep.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) doRunWithRetries(f func(int, error, Runner)) {
	const retries = 2
	expectedErr := errors.New("expected")
	sleep := new(mockSleep)
	r := suite.newRunner(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
	)

	r.(*runner).sleep = sleep.Sleep
	sleep.Expect(5 * time.Second).Times(retries)
	f(retries, expectedErr, r)

	sleep.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) doRunWithRetriesFull(f func(int, error, Runner)) {
	const retries = 2
	expectedErr := errors.New("expected")
	sleep := new(mockSleep)
	shouldRetry := new(mockShouldRetry)
	onFail := new(mockOnFail)
	r := suite.newRunner(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
		WithShouldRetry(shouldRetry.ShouldRetry),
		WithOnFail(onFail.OnFail),
	)

	r.(*runner).sleep = sleep.Sleep
	sleep.Expect(5 * time.Second).Times(2)
	shouldRetry.Expect(expectedErr, true).Times(2)
	shouldRetry.Expect(expectedErr, false).Once()
	onFail.Expect(expectedErr, 0).Once()               // initial attempt
	onFail.Expect(expectedErr, 5*time.Second).Times(2) // each retry
	f(retries, expectedErr, r)

	sleep.AssertExpectations(suite.T())
	shouldRetry.AssertExpectations(suite.T())
	onFail.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) testRunCtxCancel() {
	expectedCtx, cancel := suite.testContext()
	defer cancel()
	expectedErr := errors.New("expected")
	sleep := new(mockSleep)
	task := new(mockTask)
	r := suite.newRunner(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
	)

	task.ExpectCtx(expectedCtx, expectedErr).Once().Run(func(mock.Arguments) {
		cancel() // simulate the context being canceled while this attempt happens
	})

	actualErr := r.RunCtx(expectedCtx, task.DoCtx)

	select {
	case <-expectedCtx.Done():
		suite.Same(expectedCtx.Err(), actualErr)

	case <-time.After(time.Second):
		suite.Fail("context was not canceled")
	}

	sleep.AssertExpectations(suite.T())
	task.AssertExpectations(suite.T())
}

func (suite *RunnerSuite) TestRun() {
	suite.Run("FirstAttemptSuccess", func() {
		suite.doRunFirstAttemptSuccess(func(r Runner) {
			task := new(mockTask)
			task.Expect(nil).Once()
			suite.NoError(r.Run(task.Do))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("NoPolicyFactory", func() {
		suite.doRunNoPolicyFactory(func(expectedErr error, r Runner) {
			task := new(mockTask)
			task.Expect(expectedErr).Once()
			suite.Same(expectedErr, r.Run(task.Do))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetries", func() {
		suite.doRunWithRetries(func(retries int, expectedErr error, r Runner) {
			task := new(mockTask)
			task.Expect(expectedErr).Times(retries)
			task.Expect(nil).Once() // success
			suite.NoError(r.Run(task.Do))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetriesFull", func() {
		suite.doRunWithRetriesFull(func(retries int, expectedErr error, r Runner) {
			task := new(mockTask)
			task.Expect(expectedErr).Times(retries + 1) // all attempts fail
			suite.Same(expectedErr, r.Run(task.Do))
			task.AssertExpectations(suite.T())
		})
	})
}

func (suite *RunnerSuite) TestRunCtx() {
	suite.Run("FirstAttemptSuccess", func() {
		suite.doRunFirstAttemptSuccess(func(r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, nil).Once()
			suite.NoError(r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("NoPolicyFactory", func() {
		suite.doRunNoPolicyFactory(func(expectedErr error, r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, expectedErr).Once()
			suite.Same(expectedErr, r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetries", func() {
		suite.doRunWithRetries(func(retries int, expectedErr error, r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, expectedErr).Times(retries)
			task.ExpectCtx(expectedCtx, nil).Once() // success
			suite.NoError(r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetriesFull", func() {
		suite.doRunWithRetriesFull(func(retries int, expectedErr error, r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, expectedErr).Times(retries + 1) // all attempts fail
			suite.Same(expectedErr, r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("Cancel", suite.testRunCtxCancel)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}

func (suite *RunnerSuite) TestNewRunnerOptionError() {
	expectedErr := errors.New("expected")
	r, actualErr := NewRunner(func(*coreRunner) error {
		return expectedErr
	})

	suite.Error(actualErr)
	suite.Nil(r)
}

type RunnerWithDataSuite struct {
	taskSuite
}

func (suite *RunnerWithDataSuite) newRunner(o ...RunnerOption) RunnerWithData[int] {
	r, err := NewRunnerWithData[int](o...)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)
	return r
}

func (suite *RunnerWithDataSuite) TestRun() {
}

func (suite *RunnerWithDataSuite) TestRunCtx() {
}

func TestRunnerWithData(t *testing.T) {
	suite.Run(t, new(RunnerWithDataSuite))
}
