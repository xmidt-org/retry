package retry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RunnerSuite struct {
	CommonSuite
}

func (suite *RunnerSuite) testNoRetries() {
	var (
		testCtx, _ = suite.testCtx()
		taskCtx    context.Context
		task       = func(ctx context.Context) (int, error) {
			taskCtx = ctx // capture
			return 123, nil
		}

		onAttempt = new(mockOnAttempt)
		runner    = suite.newRunner(
			WithOnAttempt(onAttempt.OnAttempt),
		)
	)

	onAttempt.ExpectMatch(func(actual Attempt) bool {
		return suite.assertTestAttempt(
			Attempt{}, // since no retries, the non-context fields will be zeroes
			actual,
		)
	}).Once()

	result, err := runner.Run(testCtx, task)
	suite.Equal(123, result)
	suite.NoError(err)
	suite.Require().NotNil(taskCtx)
}

func (suite *RunnerSuite) TestRun() {
	suite.Run("NoRetries", suite.testNoRetries)
}

func TestRunner(t *testing.T) {
	suite.Run(t, new(RunnerSuite))
}
