package retry

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// NeverSuite holds unit tests for the behavior of the never policy.
type NeverSuite struct {
	CommonSuite
}

func (suite *NeverSuite) TestCancel() {
	ctx, _ := suite.testCtx()
	p := suite.requirePolicy(
		Config{}.NewPolicy(ctx),
	)

	p.Cancel()
	suite.assertStopped(p.Next())

	// idempotent
	p.Cancel()
	suite.assertStopped(p.Next())
}

func (suite *NeverSuite) TestNext() {
	ctx, _ := suite.testCtx()
	p := suite.requirePolicy(
		Config{}.NewPolicy(ctx),
	)

	suite.assertStopped(p.Next())
}

func TestNever(t *testing.T) {
	suite.Run(t, new(NeverSuite))
}
