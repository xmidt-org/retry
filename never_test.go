// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

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

	suite.Require().NotNil(p.Context())

	p.Cancel()
	suite.assertStopped(p.Next())
	suite.NotNil(p.Context().Err())

	// idempotent
	p.Cancel()
	suite.assertStopped(p.Next())
	suite.NotNil(p.Context().Err())
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
