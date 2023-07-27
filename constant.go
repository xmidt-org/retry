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
