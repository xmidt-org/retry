package retry

import (
	"time"

	"github.com/stretchr/testify/suite"
)

// CommonSuite has a few utilities that are commonly useful for
// policy unit tests in this package.
type CommonSuite struct {
	suite.Suite
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
