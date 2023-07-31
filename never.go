package retry

import (
	"time"
)

// never is the Policy that indicates retries should never be performed.
type never struct{}

func (n never) Next() (time.Duration, bool) { return 0, false }
