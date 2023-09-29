package retry

import (
	"context"
	"time"
)

// never is the Policy that indicates retries should never be performed.
// This type still carries a context, so that the Policy interface can still
// be used to cancel the context for the single attempt.
type never struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (n never) Context() context.Context {
	return n.ctx
}

func (n never) Cancel() {
	if n.cancel != nil {
		n.cancel()
		n.cancel = nil
	}
}

func (n never) Next() (time.Duration, bool) { return 0, false }
