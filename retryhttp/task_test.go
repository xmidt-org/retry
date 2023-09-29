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
