package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TaskSuite struct {
	suite.Suite
}

func (suite *TaskSuite) setSleepFail(o Options) Options {
	o.sleep = func(time.Duration) {
		suite.Fail("sleep should not have been called")
	}

	return o
}

func (suite *TaskSuite) sleep(expected time.Duration) func(time.Duration) {
	return func(d time.Duration) {
		suite.Equal(expected, d)
	}
}

func (suite *TaskSuite) succeedsAfter(err error, retries uint) func() error {
	count := uint(0)
	return func() error {
		suite.LessOrEqual(count, retries+1) // will be called initially, then once per retry
		count++
		if count >= retries+1 {
			return nil // success
		}

		return err
	}
}

func (suite *TaskSuite) simpleOptions(interval time.Duration) Options {
	return Options{
		Factory: Config{
			Interval: interval,
		},
		sleep: func(d time.Duration) {
			suite.Equal(interval, d)
		},
	}
}

func (suite *TaskSuite) onFail(expectedErr error, interval time.Duration, failures uint) func(error, time.Duration) {
	count := uint(0)
	return func(actualErr error, d time.Duration) {
		suite.Less(count, failures)
		if count == 0 {
			// the first failed call should always have 0 for the interval
			suite.Zero(d)
		} else {
			// each retry
			suite.Equal(interval, d)
		}

		count++
		suite.Same(expectedErr, actualErr)
	}
}

func (suite *TaskSuite) testRunNoRetries() {
	testCases := []struct {
		name    string
		options Options
	}{
		{
			name:    "NilPolicyFactory",
			options: Options{},
		},
		{
			name: "WithPolicyFactory",
			options: Options{
				Factory: Config{}, // no interval
			},
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.name, func() {
			suite.Run("NoOnFail", func() {
				var (
					options     = suite.setSleepFail(testCase.options)
					expectedErr = errors.New("expected")
				)

				actualErr := Run(options, func() error { return expectedErr })
				suite.Same(expectedErr, actualErr)
			})

			suite.Run("WithOnFail", func() {
				var (
					options      = suite.setSleepFail(testCase.options)
					expectedErr  = errors.New("expected")
					onFailCalled = false
				)

				options.OnFail = func(err error, d time.Duration) {
					onFailCalled = true
					suite.Same(expectedErr, err)
					suite.Zero(d)
				}

				actualErr := Run(options, func() error { return expectedErr })
				suite.Same(expectedErr, actualErr)
				suite.True(onFailCalled)
			})
		})
	}
}

func (suite *TaskSuite) testRunWithRetries() {
	suite.Run("NoOnFail", func() {
		suite.Run("SuccessFirstCall", func() {
			suite.NoError(
				Run(
					suite.simpleOptions(5*time.Second),
					suite.succeedsAfter(nil, 0),
				),
			)
		})

		suite.Run("SuccessAfterRetries", func() {
			suite.NoError(
				Run(
					suite.simpleOptions(5*time.Second),
					suite.succeedsAfter(errors.New("expected"), 3),
				),
			)
		})
		suite.Run("AlwaysFail", func() {
			expectedErr := errors.New("expected")
			suite.Same(
				expectedErr,
				Run(
					Options{
						Factory: Config{
							Interval:   5 * time.Second,
							MaxRetries: 5,
						},
						sleep: suite.sleep(5 * time.Second),
					},
					func() error { return expectedErr },
				),
			)
		})
	})

	suite.Run("WithOnFail", func() {
		suite.Run("SuccessFirstCall", func() {
			options := suite.simpleOptions(5 * time.Second)
			options.OnFail = suite.onFail(nil, 5*time.Second, 0)
			suite.NoError(
				Run(
					options,
					suite.succeedsAfter(nil, 0),
				),
			)
		})

		suite.Run("SuccessAfterRetries", func() {
			expectedErr := errors.New("expected")
			options := suite.simpleOptions(5 * time.Second)
			options.OnFail = suite.onFail(expectedErr, 5*time.Second, 3)
			suite.NoError(
				Run(
					options,
					suite.succeedsAfter(expectedErr, 3),
				),
			)
		})
	})
}

func (suite *TaskSuite) TestRun() {
	suite.Run("NoRetries", suite.testRunNoRetries)
	suite.Run("WithRetries", suite.testRunWithRetries)
}

func TestTask(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}
