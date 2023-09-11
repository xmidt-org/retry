package retry

import (
	"context"
	"time"

	"github.com/stretchr/testify/suite"
)

type contextKey struct{}

// CommonSuite has a few utilities that are commonly useful for
// policy unit tests in this package.
type CommonSuite struct {
	suite.Suite
}

func (suite *CommonSuite) testCtx() (context.Context, context.CancelFunc) {
	return context.WithCancel(
		context.WithValue(context.Background(), contextKey{}, "test"),
	)
}

func (suite *CommonSuite) assertTestCtx(ctx context.Context) bool {
	suite.Require().NotNil(ctx)
	return suite.Equal("test", ctx.Value(contextKey{}))
}

// assertTestAttempt asserts that the expected attempt matches the actual *except*
// as regards the context.  The actual.Context field is passed to assertTextCtx.
func (suite *CommonSuite) assertTestAttempt(expected, actual Attempt) bool {
	return suite.assertTestCtx(actual.Context) ||
		suite.Equal(expected.Err, actual.Err) ||
		suite.Equal(expected.Retries, actual.Retries) ||
		suite.Equal(expected.Next, actual.Next)
}

// requirePolicy halts the current test if p is nil.  The given Policy
// is returned for further testing.
func (suite *CommonSuite) requirePolicy(p Policy) Policy {
	suite.Require().NotNil(p)
	return p
}

// requireNever fails the enclosing test if p is not a never policy.  The
// never instance is returned for further testing.
func (suite *CommonSuite) requireNever(p Policy) never {
	suite.Require().IsType(never{}, p)
	return p.(never)
}

// requireConstant fails the enclosing test if p is not a constant policy.  The
// constant instance is returned for further testing.
func (suite *CommonSuite) requireConstant(p Policy) *constant {
	suite.Require().IsType((*constant)(nil), p)
	return p.(*constant)
}

// requireExponential fails the enclosing test if p is not an exponential policy.  The
// exponential instance is returned for further testing.
func (suite *CommonSuite) requireExponential(p Policy) *exponential {
	suite.Require().IsType((*exponential)(nil), p)
	return p.(*exponential)
}

// assertContinue asserts that the result from Policy.Next indicates that
// the retries should continue.  The time.Duration interval is returned
// for further testing.
func (suite *CommonSuite) assertContinue(d time.Duration, ok bool) time.Duration {
	suite.Greater(d, time.Duration(0))
	suite.True(ok)
	return d
}

// assertStopped asserts that the result from Policy.Next indicates that the
// retries should stop.
func (suite *CommonSuite) assertStopped(d time.Duration, ok bool) {
	suite.Zero(d)
	suite.False(ok)
}

func (suite *CommonSuite) newRunner(o ...RunnerOption) Runner[int] {
	runner, err := NewRunner[int](o...)
	suite.Require().NoError(err)
	suite.Require().NotNil(runner)
	return runner
}
