package retry

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ExponentialSuite struct {
	CommonSuite
}

func (suite *ExponentialSuite) testNextMaxRetriesExceeded() {
	testCtx, _ := suite.testCtx()
	p := suite.requirePolicy(
		Config{
			Interval:   5 * time.Second,
			Multiplier: 2.0,
			MaxRetries: 2,
		}.NewPolicy(testCtx),
	)

	suite.assertTestCtx(p.Context())
	suite.Equal(5*time.Second, suite.assertContinue(p.Next()))
	suite.Equal(10*time.Second, suite.assertContinue(p.Next()))
	suite.assertStopped(p.Next())
}

func (suite *ExponentialSuite) testNextMultiplierNoJitter() {
	const expectedInterval time.Duration = 5 * time.Second
	testCtx, _ := suite.testCtx()
	p := suite.requirePolicy(
		Config{
			Interval:   expectedInterval,
			Multiplier: 2.0,
		}.NewPolicy(testCtx),
	)

	suite.assertTestCtx(p.Context())

	for i := 0; i < 10; i++ {
		suite.Equal(
			time.Duration(float64(expectedInterval)*math.Exp2(float64(i))),
			suite.assertContinue(p.Next()),
		)
	}
}

func (suite *ExponentialSuite) testNextMultiplierWithJitter() {
	testCtx, _ := suite.testCtx()
	p := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:   5 * time.Second,
				Multiplier: 2.0,
				Jitter:     0.1,
			}.NewPolicy(testCtx),
		),
	)

	// a predictable random value
	p.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	suite.assertTestCtx(p.Context())

	for i := 0; i < 10; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(p.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(p.Next()))
	}
}

func (suite *ExponentialSuite) testNextMultiplierWithJitterAndMaxRetries() {
	testCtx, _ := suite.testCtx()
	p := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:   5 * time.Second,
				Multiplier: 2.0,
				Jitter:     0.1,
				MaxRetries: 5,
			}.NewPolicy(testCtx),
		),
	)

	suite.assertTestCtx(p.Context())

	// a predictable random value
	p.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	for i := 0; i < 5; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(p.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(p.Next()))
	}

	suite.assertStopped(p.Next())
}

func (suite *ExponentialSuite) testNextMultiplierWithJitterAndMaxInterval() {
	testCtx, _ := suite.testCtx()
	p := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:    5 * time.Second,
				Multiplier:  2.0,
				Jitter:      0.1,
				MaxInterval: 11 * time.Second,
			}.NewPolicy(testCtx),
		),
	)

	suite.assertTestCtx(p.Context())

	// a predictable random value
	p.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	// with 5s starting, the max will be hit in 2 iterations
	for i := 0; i < 2; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(p.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(p.Next()))
	}

	// jitter is applied after the max interval is enforced
	var (
		delta    = time.Duration(11.0 * float64(time.Second) * 0.1)
		expected = 11*time.Second - delta + time.Duration(p.rand(2*int64(delta)+1))
	)

	suite.Equal(expected, suite.assertContinue(p.Next()))

	// finally, verify that the jitterized interval is subject to the maxInterval
	p.rand = func(v int64) int64 {
		return v - 1 // max
	}

	suite.Equal(11*time.Second, suite.assertContinue(p.Next()))
}

func (suite *ExponentialSuite) TestNext() {
	suite.Run("MaxRetriesExceeded", suite.testNextMaxRetriesExceeded)
	suite.Run("MultiplierNoJitter", suite.testNextMultiplierNoJitter)
	suite.Run("MultiplierWithJitter", suite.testNextMultiplierWithJitter)
	suite.Run("MultiplierWithJitterAndMaxRetries", suite.testNextMultiplierWithJitterAndMaxRetries)
	suite.Run("MultiplierWithJitterAndMaxInterval", suite.testNextMultiplierWithJitterAndMaxInterval)
}
func TestExponential(t *testing.T) {
	suite.Run(t, new(ExponentialSuite))
}
