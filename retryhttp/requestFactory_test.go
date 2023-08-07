package retryhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testMethod = "GET"
	testURL    = "/test?foo=bar"
)

func (suite *FactoriesSuite) TestTaskBodyCloser() {
	var (
		reader = strings.NewReader("test body")
		body   = taskBodyCloser{
			TaskBody: reader,
		}
	)

	// verify Close did not affect the reader
	suite.NoError(body.Close())
	pos, err := reader.Seek(0, io.SeekCurrent)
	suite.NoError(err)
	suite.Equal(int64(0), pos)

	// idempotent
	suite.NoError(body.Close())
	pos, err = reader.Seek(0, io.SeekCurrent)
	suite.NoError(err)
	suite.Equal(int64(0), pos)
}

type FactoriesSuite struct {
	suite.Suite
}

func (suite *FactoriesSuite) expectedCtx() context.Context {
	type contextKey struct{}
	return context.WithValue(context.Background(), contextKey{}, "value")
}

// assertRequest runs standard assertions on our typical test request
func (suite *FactoriesSuite) assertRequest(request *http.Request, err error) *http.Request {
	suite.NoError(err)
	suite.Require().NotNil(request)

	suite.Equal(testMethod, request.Method)
	suite.Require().NotNil(request.URL)
	suite.Equal(testURL, request.URL.String())

	return request
}

// assertBody asserts that the request's body-related attributes are set correctly.
func (suite *FactoriesSuite) assertBody(expected []byte, request *http.Request) {
	suite.Require().NotNil(request.Body)
	actual, err := io.ReadAll(request.Body)
	suite.NoError(err)
	suite.Equal(expected, actual)
	suite.Equal(int64(len(expected)), request.ContentLength)
}

// assertGetBody asserts that the request's GetBody is idempotent and returns the correct body.
func (suite *FactoriesSuite) assertGetBody(expected []byte, request *http.Request) {
	suite.Require().NotNil(request.GetBody)
	for i := 0; i < 3; i++ {
		suite.T().Logf("body #%d", i)
		newBody, err := request.GetBody()
		suite.NoError(err)
		suite.Require().NotNil(newBody)

		actual, err := io.ReadAll(newBody)
		suite.NoError(err)
		suite.Equal(expected, actual)
	}
}

func (suite *FactoriesSuite) testNewRequestInvalid() {
	factory := NewRequest("NOTAVALIDMETHOD@^$*!(@*!", "this is not a valid URL", nil, nil)
	suite.Require().NotNil(factory)

	request, err := factory(context.Background())
	suite.Error(err)
	suite.Nil(request)
}

func (suite *FactoriesSuite) testNewRequestNoHeader() {
	var (
		expectedCtx = suite.expectedCtx()
		factory     = NewRequest("GET", "/test?foo=bar", strings.NewReader("test body"), nil)
	)

	suite.Require().NotNil(factory)
	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request := suite.assertRequest(factory(expectedCtx))
			suite.assertBody([]byte("test body"), request)
			suite.assertGetBody([]byte("test body"), request)
		})
	}
}

func (suite *FactoriesSuite) testNewRequestWithHeader() {
	var (
		expectedCtx    = suite.expectedCtx()
		expectedHeader = http.Header{
			"Test": []string{"true"},
		}

		factory = NewRequest("GET", "/test?foo=bar", strings.NewReader("test body"), expectedHeader)
	)

	suite.Require().NotNil(factory)
	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request := suite.assertRequest(factory(expectedCtx))
			suite.assertBody([]byte("test body"), request)
			suite.assertGetBody([]byte("test body"), request)
			suite.Equal(expectedHeader, request.Header)
		})
	}
}

func (suite *FactoriesSuite) TestNewRequest() {
	suite.Run("Invalid", suite.testNewRequestInvalid)
	suite.Run("NoHeader", suite.testNewRequestNoHeader)
	suite.Run("WithHeader", suite.testNewRequestWithHeader)
}

func (suite *FactoriesSuite) TestPrototype() {
	var (
		expectedCtx = suite.expectedCtx()
		prototype   = httptest.NewRequest(testMethod, testURL, nil)
		factory     = Prototype(prototype)
	)

	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request := suite.assertRequest(factory(expectedCtx))
			suite.assertBody([]byte{}, request)
			suite.Nil(request.GetBody)
		})
	}
}

func (suite *FactoriesSuite) TestPrototypeBytes() {
	var (
		expectedCtx = suite.expectedCtx()
		body        = []byte("test body")
		prototype   = httptest.NewRequest(testMethod, testURL, nil)
		factory     = PrototypeBytes(prototype, body)
	)

	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request := suite.assertRequest(factory(expectedCtx))
			suite.assertBody([]byte("test body"), request)
			suite.assertGetBody([]byte("test body"), request)
		})
	}
}

func (suite *FactoriesSuite) TestPrototypeReader() {
	var (
		expectedCtx = suite.expectedCtx()
		body        = bytes.NewReader([]byte("test body"))
		prototype   = httptest.NewRequest(testMethod, testURL, nil)
		factory     = PrototypeReader(prototype, body)
	)

	for i := 0; i < 3; i++ {
		suite.Run(fmt.Sprintf("iteration-%d", i), func() {
			request := suite.assertRequest(factory(expectedCtx))
			suite.assertBody([]byte("test body"), request)
			suite.assertGetBody([]byte("test body"), request)
		})
	}
}

func TestFactories(t *testing.T) {
	suite.Run(t, new(FactoriesSuite))
}
