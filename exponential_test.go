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
	p := suite.requirePolicy(
		Config{
			Interval:   5 * time.Second,
			Multiplier: 2.0,
			MaxRetries: 2,
		}.NewPolicy(),
	)

	suite.Equal(5*time.Second, suite.assertContinue(p.Next()))
	suite.Equal(10*time.Second, suite.assertContinue(p.Next()))
	suite.assertStopped(p.Next())
}

func (suite *ExponentialSuite) testNextMaxElapsedTimeExceeded() {
	e := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:       5 * time.Second,
				MaxElapsedTime: 15 * time.Second,
				Multiplier:     2.0,
			}.NewPolicy(),
		),
	)

	e.now = func() time.Time {
		return e.start.Add(15 * time.Second)
	}

	suite.assertStopped(e.Next())
}

func (suite *ExponentialSuite) testNextMultiplierNoJitter() {
	const expectedInterval time.Duration = 5 * time.Second
	p := suite.requirePolicy(
		Config{
			Interval:   expectedInterval,
			Multiplier: 2.0,
		}.NewPolicy(),
	)

	for i := 0; i < 10; i++ {
		suite.Equal(
			time.Duration(float64(expectedInterval)*math.Exp2(float64(i))),
			suite.assertContinue(p.Next()),
		)
	}
}

func (suite *ExponentialSuite) testNextMultiplierWithJitter() {
	e := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:   5 * time.Second,
				Multiplier: 2.0,
				Jitter:     0.1,
			}.NewPolicy(),
		),
	)

	// a predictable random value
	e.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	for i := 0; i < 10; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(e.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(e.Next()))
	}
}

func (suite *ExponentialSuite) testNextMultiplierWithJitterAndMaxRetries() {
	e := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:   5 * time.Second,
				Multiplier: 2.0,
				Jitter:     0.1,
				MaxRetries: 5,
			}.NewPolicy(),
		),
	)

	// a predictable random value
	e.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	for i := 0; i < 5; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(e.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(e.Next()))
	}

	suite.assertStopped(e.Next())
}

func (suite *ExponentialSuite) testNextMultiplierWithJitterAndMaxElapsedTime() {
	e := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:       5 * time.Second,
				Multiplier:     2.0,
				Jitter:         0.1,
				MaxElapsedTime: 15 * time.Minute,
			}.NewPolicy(),
		),
	)

	// a predictable random value
	e.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	for i := 0; i < 5; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(e.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(e.Next()))
	}

	e.now = func() time.Time {
		return e.start.Add(15 * time.Minute)
	}

	suite.assertStopped(e.Next())
}

func (suite *ExponentialSuite) testNextMultiplierWithJitterAndMaxInterval() {
	e := suite.requireExponential(
		suite.requirePolicy(
			Config{
				Interval:    5 * time.Second,
				Multiplier:  2.0,
				Jitter:      0.1,
				MaxInterval: 11 * time.Second,
			}.NewPolicy(),
		),
	)

	// a predictable random value
	e.rand = func(v int64) int64 {
		return int64(0.25 * float64(v))
	}

	// with 5s starting, the max will be hit in 2 iterations
	for i := 0; i < 2; i++ {
		var (
			base     = 5 * time.Second * time.Duration(math.Pow(2.0, float64(i)))
			delta    = time.Duration(float64(base) * 0.1)
			expected = base - delta + time.Duration(e.rand(2*int64(delta)+1))
		)

		suite.Equal(expected, suite.assertContinue(e.Next()))
	}

	// jitter is applied after the max interval is enforced
	var (
		delta    = time.Duration(11.0 * float64(time.Second) * 0.1)
		expected = 11*time.Second - delta + time.Duration(e.rand(2*int64(delta)+1))
	)

	suite.Equal(expected, suite.assertContinue(e.Next()))

	// finally, verify that the jitterized interval is subject to the maxInterval
	e.rand = func(v int64) int64 {
		return v - 1 // max
	}

	suite.Equal(11*time.Second, suite.assertContinue(e.Next()))
}

func (suite *ExponentialSuite) TestNext() {
	suite.Run("MaxRetriesExceeded", suite.testNextMaxRetriesExceeded)
	suite.Run("MaxElapsedTimeExceeded", suite.testNextMaxElapsedTimeExceeded)
	suite.Run("MultiplierNoJitter", suite.testNextMultiplierNoJitter)
	suite.Run("MultiplierWithJitter", suite.testNextMultiplierWithJitter)
	suite.Run("MultiplierWithJitterAndMaxRetries", suite.testNextMultiplierWithJitterAndMaxRetries)
	suite.Run("MultiplierWithJitterAndMaxElapsedTime", suite.testNextMultiplierWithJitterAndMaxElapsedTime)
	suite.Run("MultiplierWithJitterAndMaxInterval", suite.testNextMultiplierWithJitterAndMaxInterval)
}
func TestExponential(t *testing.T) {
	suite.Run(t, new(ExponentialSuite))
}
