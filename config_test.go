// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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
	testCtx, _ := suite.testCtx()
	p := suite.requireNever(
		suite.requirePolicy(
			Config{}.NewPolicy(testCtx),
		),
	)

	suite.NotNil(p.ctx)
	suite.NotNil(p.cancel)
}

func (suite *ConfigSuite) TestConstant() {
	testCtx, _ := suite.testCtx()
	p := suite.requireConstant(
		suite.requirePolicy(
			Config{
				MaxRetries:     5,
				MaxElapsedTime: 5 * time.Hour,
				Interval:       27 * time.Second,
			}.NewPolicy(testCtx),
		),
	)

	suite.NotNil(p.ctx)
	suite.NotNil(p.cancel)
	suite.Equal(5, p.maxRetries)
	suite.Equal(27*time.Second, p.interval)

	deadline, ok := p.ctx.Deadline()
	suite.Require().True(ok)
	suite.GreaterOrEqual(
		time.Now().Add(5*time.Hour),
		deadline,
	)
}

func (suite *ConfigSuite) TestExponential() {
	testCtx, _ := suite.testCtx()
	p := suite.requireExponential(
		suite.requirePolicy(
			Config{
				MaxRetries:     5,
				MaxElapsedTime: 5 * time.Hour,
				Interval:       6 * time.Second,
				Jitter:         0.1,
				Multiplier:     2.0,
				MaxInterval:    15 * time.Hour,
			}.NewPolicy(testCtx),
		),
	)

	suite.Equal(5, p.maxRetries)
	suite.Equal(6*time.Second, p.initial)
	suite.Zero(p.previous)
	suite.Equal(0.1, p.jitter)
	suite.Equal(2.0, p.multiplier)
	suite.Equal(15*time.Hour, p.maxInterval)

	deadline, ok := p.ctx.Deadline()
	suite.Require().True(ok)
	suite.GreaterOrEqual(
		time.Now().Add(5*time.Hour),
		deadline,
	)
}

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
