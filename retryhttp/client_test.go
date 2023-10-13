package retryhttp

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ClientSuite struct {
	suite.Suite
}

func (suite *ClientSuite) TestDo() {
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
