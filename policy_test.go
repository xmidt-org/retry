package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PolicySuite struct {
	CommonSuite
}

func (suite *PolicySuite) TestPolicyFactoryFunc() {
	var (
		expected = &constant{
			interval: 16 * time.Second,
		}

		pf     PolicyFactory = PolicyFactoryFunc(func() Policy { return expected })
		actual               = pf.NewPolicy()
	)

	suite.Same(expected, actual)
}

func TestPolicy(t *testing.T) {
	suite.Run(t, new(PolicySuite))
}
