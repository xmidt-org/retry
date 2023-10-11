package retryhttp

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/retry"
)

type NewShouldRetrySuite struct {
	suite.Suite
}

func (suite *NewShouldRetrySuite) newShouldRetry(statusCodes ...int) retry.ShouldRetry[*http.Response] {
	sr := NewShouldRetry(statusCodes...)
	suite.Require().NotNil(sr)
	return sr
}

func (suite *NewShouldRetrySuite) testTemporaryError(sr retry.ShouldRetry[*http.Response]) {
	suite.Run("NilResponse", func() {
		for _, temporary := range []bool{true, false} {
			suite.Run(fmt.Sprintf("Temporary=%t", temporary), func() {
				suite.Equal(
					temporary,
					sr(nil, &net.DNSError{IsTemporary: temporary}),
				)
			})
		}
	})

	suite.Run("WithResponse", func() {
		for _, temporary := range []bool{true, false} {
			suite.Run(fmt.Sprintf("Temporary=%t", temporary), func() {
				response := &http.Response{
					StatusCode: http.StatusTooManyRequests,
				}

				suite.Equal(
					temporary,
					sr(response, &net.DNSError{IsTemporary: temporary}),
				)
			})
		}
	})
}

func (suite *NewShouldRetrySuite) TestTemporaryError() {
	suite.Run("NoCodes", func() {
		sr := suite.newShouldRetry()
		suite.testTemporaryError(sr)
	})

	suite.Run("WithCodes", func() {
		sr := suite.newShouldRetry(http.StatusTooManyRequests)
		suite.testTemporaryError(sr)
	})
}

func (suite *NewShouldRetrySuite) TestFatalError() {
	suite.False(
		suite.newShouldRetry()(nil, errors.New("this would be fatal")),
	)
}

func (suite *NewShouldRetrySuite) TestStatusCodeRetry() {
	sr := suite.newShouldRetry(http.StatusTooManyRequests, http.StatusInternalServerError)
	suite.True(
		sr(&http.Response{StatusCode: http.StatusTooManyRequests}, nil),
	)

	suite.True(
		sr(&http.Response{StatusCode: http.StatusInternalServerError}, nil),
	)
}

func (suite *NewShouldRetrySuite) TestStatusCodeNoRetry() {
	sr := suite.newShouldRetry(http.StatusTooManyRequests, http.StatusInternalServerError)
	suite.False(
		sr(&http.Response{StatusCode: http.StatusOK}, nil),
	)

	suite.False(
		sr(&http.Response{StatusCode: http.StatusNotFound}, nil),
	)
}

func TestNewShouldRetry(t *testing.T) {
	suite.Run(t, new(NewShouldRetrySuite))
}
