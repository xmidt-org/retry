package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ConfigSuite holds unit tests around the structural creation of various policies.
// No behavior tests are implemented here.
type ConfigSuite struct {
	CommonSuite
}

func (suite *ConfigSuite) TestDefault() {
	suite.requireNever(
		suite.requirePolicy(
			Config{}.NewPolicy(),
		),
	)
}

func (suite *ConfigSuite) TestConstant() {
	c := suite.requireConstant(
		suite.requirePolicy(
			Config{
				MaxRetries:     5,
				MaxElapsedTime: 5 * time.Hour,
				Interval:       27 * time.Second,
			}.NewPolicy(),
		),
	)

	suite.Equal(5, c.maxRetries)
	suite.Equal(5*time.Hour, c.maxElapsedTime)
	suite.Equal(27*time.Second, c.interval)
	suite.NotNil(c.now)
	suite.GreaterOrEqual(time.Now(), c.start)
}

func (suite *ConfigSuite) TestExponential() {
	e := suite.requireExponential(
		suite.requirePolicy(
			Config{
				MaxRetries:     5,
				MaxElapsedTime: 5 * time.Hour,
				Interval:       6 * time.Second,
				Jitter:         0.1,
				Multiplier:     2.0,
				MaxInterval:    15 * time.Hour,
			}.NewPolicy(),
		),
	)

	suite.Equal(5, e.maxRetries)
	suite.Equal(5*time.Hour, e.maxElapsedTime)
	suite.Equal(6*time.Second, e.initial)
	suite.Zero(e.previous)
	suite.Equal(0.1, e.jitter)
	suite.Equal(2.0, e.multiplier)
	suite.Equal(15*time.Hour, e.maxInterval)
	suite.NotNil(e.now)
	suite.GreaterOrEqual(time.Now(), e.start)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
