// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type AttemptSuite struct {
	suite.Suite
}

func (suite *AttemptSuite) TestDone() {
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	testCases := []struct {
		attempt  Attempt[int]
		expected bool
		label    string
	}{
		{
			attempt: Attempt[int]{
				Context: context.Background(),
				Next:    5 * time.Second,
			},
			expected: false,
			label:    "NotDone",
		},
		{
			attempt: Attempt[int]{
				Context: canceledCtx,
				Next:    5 * time.Second,
			},
			expected: true,
			label:    "ContextCancelled",
		},
		{
			attempt: Attempt[int]{
				Context: context.Background(),
				Next:    0,
			},
			expected: true,
			label:    "NoNext",
		},
	}

	for _, testCase := range testCases {
		suite.Run(testCase.label, func() {
			suite.Equal(
				testCase.expected,
				testCase.attempt.Done(),
			)
		})
	}
}

func TestAttempt(t *testing.T) {
	suite.Run(t, new(AttemptSuite))
}
