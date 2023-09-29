// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TaskSuite struct {
	suite.Suite
}

func TestTask(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}
