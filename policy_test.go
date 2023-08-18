package retry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PolicySuite struct {
	CommonSuite
}

func (suite *PolicySuite) TestPolicyFactoryFunc() {
	var (
		expectedInterval = 5 * time.Second

		pf PolicyFactory = PolicyFactoryFunc(func(ctx context.Context) Policy {
			return &constant{
				corePolicy: corePolicy{
					ctx: ctx,
					// leave cancel nil, as we don't need it for this test
				},
				interval: expectedInterval,
			}
		})

		testCtx, _ = suite.testCtx()
		actual     = pf.NewPolicy(testCtx)
	)

	suite.IsType((*constant)(nil), actual)
	suite.Same(testCtx, actual.(*constant).ctx)
	suite.Equal(expectedInterval, actual.(*constant).interval)
}

func TestPolicy(t *testing.T) {
	suite.Run(t, new(PolicySuite))
}
