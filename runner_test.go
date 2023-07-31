package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// commonRunnerSuite is a generic suite with the common code for testings task runners.
type commonRunnerSuite[R any, C func(...RunnerOption) (R, *mockSleep)] struct {
	suite.Suite
	constructor C
	expectedErr error
}

func (suite *commonRunnerSuite[R, C]) SetupSuite() {
	suite.expectedErr = errors.New("expected")
}

func (suite *commonRunnerSuite[R, C]) testContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func (suite *commonRunnerSuite[R, C]) doRunFirstAttemptSuccess(f func(R)) {
	r, sleep := suite.constructor(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
	)

	f(r)
	sleep.AssertExpectations(suite.T())
}

func (suite *commonRunnerSuite[R, C]) doRunNoPolicyFactory(f func(R)) {
	r, sleep := suite.constructor()

	// no policy factory means no retries
	f(r)
	sleep.AssertExpectations(suite.T())
}

func (suite *commonRunnerSuite[R, C]) doRunWithRetries(f func(int, R)) {
	const retries = 2
	r, sleep := suite.constructor(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
	)

	sleep.Expect(5 * time.Second).Times(retries)
	f(retries, r)
	sleep.AssertExpectations(suite.T())
}

func (suite *commonRunnerSuite[R, C]) doRunWithRetriesFull(f func(int, R)) {
	const retries = 2
	shouldRetry := new(mockShouldRetry)
	onFail := new(mockOnFail)
	r, sleep := suite.constructor(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
		WithShouldRetry(shouldRetry.ShouldRetry),
		WithOnFail(onFail.OnFail),
	)

	sleep.Expect(5 * time.Second).Times(2)
	shouldRetry.Expect(suite.expectedErr, true).Times(2)
	shouldRetry.Expect(suite.expectedErr, false).Once()
	onFail.Expect(suite.expectedErr, 0).Once()               // initial attempt
	onFail.Expect(suite.expectedErr, 5*time.Second).Times(2) // each retry
	f(retries, r)

	sleep.AssertExpectations(suite.T())
	shouldRetry.AssertExpectations(suite.T())
	onFail.AssertExpectations(suite.T())
}

func (suite *commonRunnerSuite[R, C]) testRunCtxCancel(f func(context.Context, context.CancelFunc, R)) {
	expectedCtx, cancel := suite.testContext()
	defer cancel()
	r, sleep := suite.constructor(
		WithPolicyFactory(Config{Interval: 5 * time.Second}),
	)

	f(expectedCtx, cancel, r)
	suite.Error(expectedCtx.Err())
	sleep.AssertExpectations(suite.T())
}

type RunnerSuite struct {
	commonRunnerSuite[Runner, func(...RunnerOption) (Runner, *mockSleep)]
}

func (suite *RunnerSuite) SetupSuite() {
	suite.commonRunnerSuite.SetupSuite()
	suite.constructor = suite.newRunner
}

func (suite *RunnerSuite) newRunner(o ...RunnerOption) (Runner, *mockSleep) {
	r, err := NewRunner(o...)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	sleep := new(mockSleep)
	r.(*runner).sleep = sleep.Sleep

	return r, sleep
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
		suite.doRunNoPolicyFactory(func(r Runner) {
			task := new(mockTask)
			task.Expect(suite.expectedErr).Once()
			suite.Same(suite.expectedErr, r.Run(task.Do))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetries", func() {
		suite.doRunWithRetries(func(retries int, r Runner) {
			task := new(mockTask)
			task.Expect(suite.expectedErr).Times(retries)
			task.Expect(nil).Once() // success
			suite.NoError(r.Run(task.Do))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetriesFull", func() {
		suite.doRunWithRetriesFull(func(retries int, r Runner) {
			task := new(mockTask)
			task.Expect(suite.expectedErr).Times(retries + 1) // all attempts fail
			suite.Same(suite.expectedErr, r.Run(task.Do))
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
		suite.doRunNoPolicyFactory(func(r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, suite.expectedErr).Once()
			suite.Same(suite.expectedErr, r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetries", func() {
		suite.doRunWithRetries(func(retries int, r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, suite.expectedErr).Times(retries)
			task.ExpectCtx(expectedCtx, nil).Once() // success
			suite.NoError(r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetriesFull", func() {
		suite.doRunWithRetriesFull(func(retries int, r Runner) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTask)
			task.ExpectCtx(expectedCtx, suite.expectedErr).Times(retries + 1) // all attempts fail
			suite.Same(suite.expectedErr, r.RunCtx(expectedCtx, task.DoCtx))
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("Cancel", func() {
		suite.testRunCtxCancel(func(expectedCtx context.Context, cancel context.CancelFunc, r Runner) {
			task := new(mockTask)

			// failed task, and simulate the context being canceled during execution
			task.ExpectCtx(expectedCtx, suite.expectedErr).Once().Run(func(mock.Arguments) {
				cancel()
			})

			actualErr := r.RunCtx(expectedCtx, task.DoCtx)
			suite.Same(expectedCtx.Err(), actualErr)
			task.AssertExpectations(suite.T())
		})
	})
}

func (suite *RunnerSuite) TestNewRunnerOptionError() {
	r, err := NewRunner(func(*coreRunner) error {
		return suite.expectedErr
	})

	suite.Error(err)
	suite.Nil(r)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}

type RunnerWithDataSuite struct {
	commonRunnerSuite[RunnerWithData[int], func(...RunnerOption) (RunnerWithData[int], *mockSleep)]
}

func (suite *RunnerWithDataSuite) SetupSuite() {
	suite.commonRunnerSuite.SetupSuite()
	suite.constructor = suite.newRunnerWithData
}

func (suite *RunnerWithDataSuite) newRunnerWithData(o ...RunnerOption) (RunnerWithData[int], *mockSleep) {
	r, err := NewRunnerWithData[int](o...)
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	sleep := new(mockSleep)
	r.(*runnerWithData[int]).sleep = sleep.Sleep

	return r, sleep
}

func (suite *RunnerWithDataSuite) TestRun() {
	suite.Run("FirstAttemptSuccess", func() {
		suite.doRunFirstAttemptSuccess(func(r RunnerWithData[int]) {
			task := new(mockTaskWithData[int])
			task.Expect(123, nil).Once()
			result, err := r.Run(task.Do)
			suite.Equal(123, result)
			suite.NoError(err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("NoPolicyFactory", func() {
		suite.doRunNoPolicyFactory(func(r RunnerWithData[int]) {
			task := new(mockTaskWithData[int])
			task.Expect(0, suite.expectedErr).Once()
			result, err := r.Run(task.Do)
			suite.Zero(result)
			suite.Same(suite.expectedErr, err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetries", func() {
		suite.doRunWithRetries(func(retries int, r RunnerWithData[int]) {
			task := new(mockTaskWithData[int])
			task.Expect(0, suite.expectedErr).Times(retries)
			task.Expect(123, nil).Once() // success
			result, err := r.Run(task.Do)
			suite.Equal(123, result)
			suite.NoError(err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetriesFull", func() {
		suite.doRunWithRetriesFull(func(retries int, r RunnerWithData[int]) {
			task := new(mockTaskWithData[int])
			task.Expect(0, suite.expectedErr).Times(retries + 1) // all attempts fail
			result, err := r.Run(task.Do)
			suite.Zero(result)
			suite.Same(suite.expectedErr, err)
			task.AssertExpectations(suite.T())
		})
	})
}

func (suite *RunnerWithDataSuite) TestRunCtx() {
	suite.Run("FirstAttemptSuccess", func() {
		suite.doRunFirstAttemptSuccess(func(r RunnerWithData[int]) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTaskWithData[int])
			task.ExpectCtx(expectedCtx, 123, nil).Once()
			result, err := r.RunCtx(expectedCtx, task.DoCtx)
			suite.Equal(123, result)
			suite.NoError(err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("NoPolicyFactory", func() {
		suite.doRunNoPolicyFactory(func(r RunnerWithData[int]) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTaskWithData[int])
			task.ExpectCtx(expectedCtx, -1, suite.expectedErr).Once()
			result, err := r.RunCtx(expectedCtx, task.DoCtx)
			suite.Equal(-1, result)
			suite.Same(suite.expectedErr, err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetries", func() {
		suite.doRunWithRetries(func(retries int, r RunnerWithData[int]) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTaskWithData[int])
			task.ExpectCtx(expectedCtx, 0, suite.expectedErr).Times(retries)
			task.ExpectCtx(expectedCtx, 123, nil).Once() // success
			result, err := r.RunCtx(expectedCtx, task.DoCtx)
			suite.Equal(123, result)
			suite.NoError(err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("WithRetriesFull", func() {
		suite.doRunWithRetriesFull(func(retries int, r RunnerWithData[int]) {
			expectedCtx, cancel := suite.testContext()
			defer cancel()
			task := new(mockTaskWithData[int])
			task.ExpectCtx(expectedCtx, 0, suite.expectedErr).Times(retries + 1) // all attempts fail
			result, err := r.RunCtx(expectedCtx, task.DoCtx)
			suite.Zero(result)
			suite.Same(suite.expectedErr, err)
			task.AssertExpectations(suite.T())
		})
	})

	suite.Run("Cancel", func() {
		suite.testRunCtxCancel(func(expectedCtx context.Context, cancel context.CancelFunc, r RunnerWithData[int]) {
			task := new(mockTaskWithData[int])

			// failed task, and simulate the context being canceled during execution
			task.ExpectCtx(expectedCtx, -1, suite.expectedErr).Once().Run(func(mock.Arguments) {
				cancel()
			})

			result, actualErr := r.RunCtx(expectedCtx, task.DoCtx)
			suite.Equal(-1, result)
			suite.Same(expectedCtx.Err(), actualErr)
			task.AssertExpectations(suite.T())
		})
	})
}

func (suite *RunnerWithDataSuite) TestNewRunnerWithDataOptionError() {
	r, err := NewRunnerWithData[int](func(*coreRunner) error {
		return suite.expectedErr
	})

	suite.Error(err)
	suite.Nil(r)
}

func TestRunnerWithData(t *testing.T) {
	suite.Run(t, new(RunnerWithDataSuite))
}
