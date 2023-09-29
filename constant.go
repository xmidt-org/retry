// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import "time"

type constant struct {
	corePolicy
	interval time.Duration
}

func (c *constant) Next() (time.Duration, bool) {
	if !c.withinLimits() {
		return 0, false
	}

	c.retryCount++
	return c.interval, true
}
