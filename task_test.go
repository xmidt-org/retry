// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TaskSuite struct {
	suite.Suite
}

func (suite *TaskSuite) taskCtx() context.Context {
	type contextKey struct{}
	return context.WithValue(context.Background(), contextKey{}, "value")
}

func (suite *TaskSuite) testAsTaskNoContext() {
	suite.Run("NoError", func() {
		task := AsTask[int](func() (int, error) { return 123, nil })
		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.NoError(actualErr)
		suite.Equal(123, actual)
	})

	suite.Run("WithError", func() {
		var (
			expectedErr = errors.New("expected")
			task        = AsTask[int](func() (int, error) { return 123, expectedErr })
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.Same(expectedErr, actualErr)
		suite.Equal(123, actual)
	})
}

func (suite *TaskSuite) testAsTaskWithContext() {
	suite.Run("NoError", func() {
		var (
			expectedCtx = suite.taskCtx()
			task        = AsTask[int](func(actualCtx context.Context) (int, error) {
				suite.Same(expectedCtx, actualCtx)
				return 123, nil
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.NoError(actualErr)
		suite.Equal(123, actual)
	})

	suite.Run("WithError", func() {
		var (
			expectedCtx = suite.taskCtx()
			expectedErr = errors.New("expected")
			task        = AsTask[int](func(actualCtx context.Context) (int, error) {
				suite.Same(expectedCtx, actualCtx)
				return 123, expectedErr
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.Same(expectedErr, actualErr)
		suite.Equal(123, actual)
	})
}

func (suite *TaskSuite) TestAsTask() {
	suite.Run("NoContext", suite.testAsTaskNoContext)
	suite.Run("WithContext", suite.testAsTaskWithContext)
}

func (suite *TaskSuite) testAddZeroNoContext() {
	suite.Run("NoError", func() {
		task := AddZero[int](func() error { return nil })
		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.NoError(actualErr)
		suite.Equal(0, actual)
	})

	suite.Run("WithError", func() {
		var (
			expectedErr = errors.New("expected")
			task        = AddZero[int](func() error { return expectedErr })
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.Same(expectedErr, actualErr)
		suite.Equal(0, actual)
	})
}

func (suite *TaskSuite) testAddZeroWithContext() {
	suite.Run("NoError", func() {
		var (
			expectedCtx = suite.taskCtx()
			task        = AddZero[int](func(actualCtx context.Context) error {
				suite.Same(expectedCtx, actualCtx)
				return nil
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.NoError(actualErr)
		suite.Equal(0, actual)
	})

	suite.Run("WithError", func() {
		var (
			expectedCtx = suite.taskCtx()
			expectedErr = errors.New("expected")
			task        = AddZero[int](func(actualCtx context.Context) error {
				suite.Same(expectedCtx, actualCtx)
				return expectedErr
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.Same(expectedErr, actualErr)
		suite.Equal(0, actual)
	})
}

func (suite *TaskSuite) TestAddZero() {
	suite.Run("NoContext", suite.testAddZeroNoContext)
	suite.Run("WithContext", suite.testAddZeroWithContext)
}

func (suite *TaskSuite) testAddValueNoContext() {
	suite.Run("NoError", func() {
		const expected = 123
		task := AddValue(expected, func() error { return nil })
		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.NoError(actualErr)
		suite.Equal(expected, actual)
	})

	suite.Run("WithError", func() {
		const expected = 123

		var (
			expectedErr = errors.New("expected")
			task        = AddValue(expected, func() error { return expectedErr })
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.Same(expectedErr, actualErr)
		suite.Equal(expected, actual)
	})
}

func (suite *TaskSuite) testAddValueWithContext() {
	suite.Run("NoError", func() {
		const expected = 123

		var (
			expectedCtx = suite.taskCtx()
			task        = AddValue(expected, func(actualCtx context.Context) error {
				suite.Same(expectedCtx, actualCtx)
				return nil
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.NoError(actualErr)
		suite.Equal(expected, actual)
	})

	suite.Run("WithError", func() {
		const expected = 123

		var (
			expectedCtx = suite.taskCtx()
			expectedErr = errors.New("expected")
			task        = AddValue(expected, func(actualCtx context.Context) error {
				suite.Same(expectedCtx, actualCtx)
				return expectedErr
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.Same(expectedErr, actualErr)
		suite.Equal(expected, actual)
	})
}

func (suite *TaskSuite) TestAddValue() {
	suite.Run("NoContext", suite.testAddValueNoContext)
	suite.Run("WithContext", suite.testAddValueWithContext)
}

func (suite *TaskSuite) testAddSuccessFailNoContext() {
	suite.Run("NoError", func() {
		const success, fail = 123, 456
		task := AddSuccessFail(success, fail, func() error { return nil })
		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.NoError(actualErr)
		suite.Equal(success, actual)
	})

	suite.Run("WithError", func() {
		const success, fail = 123, 456

		var (
			expectedErr = errors.New("expected")
			task        = AddSuccessFail(success, fail, func() error { return expectedErr })
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(context.Background())
		suite.Same(expectedErr, actualErr)
		suite.Equal(fail, actual)
	})
}

func (suite *TaskSuite) testAddSuccessFailWithContext() {
	suite.Run("NoError", func() {
		const success, fail = 123, 456

		var (
			expectedCtx = suite.taskCtx()
			task        = AddSuccessFail(success, fail, func(actualCtx context.Context) error {
				suite.Same(expectedCtx, actualCtx)
				return nil
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.NoError(actualErr)
		suite.Equal(success, actual)
	})

	suite.Run("WithError", func() {
		const success, fail = 123, 456

		var (
			expectedCtx = suite.taskCtx()
			expectedErr = errors.New("expected")
			task        = AddSuccessFail(success, fail, func(actualCtx context.Context) error {
				suite.Same(expectedCtx, actualCtx)
				return expectedErr
			})
		)

		suite.Require().NotNil(task)
		actual, actualErr := task(expectedCtx)
		suite.Same(expectedErr, actualErr)
		suite.Equal(fail, actual)
	})
}

func (suite *TaskSuite) TestAddSuccessFail() {
	suite.Run("NoContext", suite.testAddSuccessFailNoContext)
	suite.Run("WithContext", suite.testAddSuccessFailWithContext)
}

func TestTask(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}
