package retryhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FactoriesSuite struct {
	suite.Suite
}

func (suite *FactoriesSuite) expectedCtx() context.Context {
	type contextKey struct{}
	return context.WithValue(context.Background(), contextKey{}, "value")
}

func (suite *FactoriesSuite) TestPrototype() {
	var (
		expectedCtx = suite.expectedCtx()
		prototype   = httptest.NewRequest("GET", "/test?foo=bar", nil)
		factory     = Prototype(prototype)
	)

	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request, err := factory(expectedCtx)
			suite.NoError(err)
			suite.Require().NotNil(request)
			suite.Equal("/test?foo=bar", request.URL.String())
			suite.Equal("GET", request.Method)
			suite.Empty(request.Body)
		})
	}
}

func (suite *FactoriesSuite) TestPrototypeBytes() {
	var (
		expectedCtx = suite.expectedCtx()
		body        = []byte("test body")
		prototype   = httptest.NewRequest("GET", "/test?foo=bar", nil)
		factory     = PrototypeBytes(prototype, body)
	)

	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request, err := factory(expectedCtx)
			suite.NoError(err)
			suite.Require().NotNil(request)
			suite.Equal("/test?foo=bar", request.URL.String())
			suite.Equal("GET", request.Method)

			actual, err := io.ReadAll(request.Body)
			suite.NoError(err)
			suite.Equal(body, actual)

			suite.Require().NotNil(request.GetBody)

			// make sure GetBody is idempotent
			for i := 0; i < 3; i++ {
				suite.Run(fmt.Sprintf("getBody-%d", i), func() {
					newBody, err := request.GetBody()
					suite.NoError(err)
					suite.Require().NotNil(newBody)

					actual, err := io.ReadAll(newBody)
					suite.NoError(err)
					suite.Equal(body, actual)
				})
			}
		})
	}
}

func (suite *FactoriesSuite) TestPrototypeReader() {
	var (
		expectedCtx = suite.expectedCtx()
		body        = bytes.NewReader([]byte("test body"))
		prototype   = httptest.NewRequest("GET", "/test?foo=bar", nil)
		factory     = PrototypeReader(prototype, body)
	)

	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request, err := factory(expectedCtx)
			suite.NoError(err)
			suite.Require().NotNil(request)
			suite.Equal("/test?foo=bar", request.URL.String())
			suite.Equal("GET", request.Method)

			actual, err := io.ReadAll(request.Body)
			suite.NoError(err)
			suite.Equal("test body", string(actual))

			suite.Require().NotNil(request.GetBody)

			// make sure GetBody is idempotent
			for i := 0; i < 3; i++ {
				suite.Run(fmt.Sprintf("getBody-%d", i), func() {
					newBody, err := request.GetBody()
					suite.NoError(err)
					suite.Require().NotNil(newBody)

					actual, err := io.ReadAll(newBody)
					suite.NoError(err)
					suite.Equal("test body", string(actual))
				})
			}
		})
	}
}

func TestFactories(t *testing.T) {
	suite.Run(t, new(FactoriesSuite))
}
