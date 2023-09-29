// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type ConstantSuite struct {
	CommonSuite
}

func (suite *ConstantSuite) testNextMaxRetriesExceeded() {
	testCtx, _ := suite.testCtx()
	p := suite.requirePolicy(
		Config{
			Interval:   5 * time.Second,
			MaxRetries: 2,
		}.NewPolicy(testCtx),
	)

	suite.assertTestCtx(p.Context())
	suite.Equal(5*time.Second, suite.assertContinue(p.Next()))
	suite.Equal(5*time.Second, suite.assertContinue(p.Next()))
	suite.assertStopped(p.Next())
}

func (suite *ConstantSuite) testNextNoLimit() {
	testCtx, _ := suite.testCtx()
	p := suite.requirePolicy(
		Config{
			Interval: 5 * time.Second,
		}.NewPolicy(testCtx),
	)

	suite.assertTestCtx(p.Context())

	for i := 0; i < 10; i++ {
		suite.Equal(
			5*time.Second,
			suite.assertContinue(p.Next()),
		)
	}
}

func (suite *ConstantSuite) testNextCancel() {
	testCtx, cancel := suite.testCtx()
	p := suite.requirePolicy(
		Config{
			Interval: 5 * time.Second,
		}.NewPolicy(testCtx),
	)

	suite.assertTestCtx(p.Context())

	cancel()
	suite.assertStopped(p.Next())
}

func (suite *ConstantSuite) TestNext() {
	suite.Run("MaxRetriesExceeded", suite.testNextMaxRetriesExceeded)
	suite.Run("NoLimit", suite.testNextNoLimit)
	suite.Run("Cancel", suite.testNextCancel)
}

func TestConstant(t *testing.T) {
	suite.Run(t, new(ConstantSuite))
}
