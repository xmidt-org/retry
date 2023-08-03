package retryhttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/retry"
)

type TaskSuite struct {
	suite.Suite
}

func (suite *TaskSuite) expectedCtx() context.Context {
	type contextKey struct{}
	return context.WithValue(context.Background(), contextKey{}, "value")
}

func (suite *TaskSuite) testDoCtxNoFactory() {
	suite.Panics(func() {
		Task[bool]{}.DoCtx(suite.expectedCtx())
	})
}

func (suite *TaskSuite) testDoCtxFactoryError() {
	var (
		factory   = new(mockRequestFactory)
		client    = new(mockClient)
		converter = new(mockConverter[int])

		expectedCtx     = suite.expectedCtx()
		expectedErr     = errors.New("expected")
		expectedRequest = httptest.NewRequest("GET", "/", nil)

		task = Task[int]{
			Factory:   factory.New,
			Client:    client.Do,
			Converter: converter.Convert,
		}
	)

	r, err := retry.NewRunnerWithData[int]()
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	factory.ExpectNew(expectedCtx, expectedRequest, expectedErr)

	actual, actualErr := r.RunCtx(expectedCtx, task.DoCtx)
	suite.Zero(actual)
	suite.Same(expectedErr, actualErr)

	factory.AssertExpectations(suite.T())
	client.AssertExpectations(suite.T())
	converter.AssertExpectations(suite.T())
}

func (suite *TaskSuite) testDoCtxClientError() {
	var (
		factory   = new(mockRequestFactory)
		client    = new(mockClient)
		converter = new(mockConverter[int])

		expectedCtx     = suite.expectedCtx()
		expectedErr     = errors.New("expected")
		expectedRequest = httptest.NewRequest("GET", "/", nil)

		task = Task[int]{
			Factory:   factory.New,
			Client:    client.Do,
			Converter: converter.Convert,
		}
	)

	r, err := retry.NewRunnerWithData[int]()
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	factory.ExpectNew(expectedCtx, expectedRequest, nil)
	client.ExpectDo(expectedRequest, nil, expectedErr)

	actual, actualErr := r.RunCtx(expectedCtx, task.DoCtx)
	suite.Zero(actual)
	suite.Same(expectedErr, actualErr)

	factory.AssertExpectations(suite.T())
	client.AssertExpectations(suite.T())
	converter.AssertExpectations(suite.T())
}

func (suite *TaskSuite) testDoCtxConverterError() {
	var (
		factory   = new(mockRequestFactory)
		client    = new(mockClient)
		converter = new(mockConverter[int])

		expectedCtx      = suite.expectedCtx()
		expectedErr      = errors.New("expected")
		expectedRequest  = httptest.NewRequest("GET", "/", nil)
		expectedResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(new(bytes.Buffer)),
		}

		task = Task[int]{
			Factory:   factory.New,
			Client:    client.Do,
			Converter: converter.Convert,
		}
	)

	r, err := retry.NewRunnerWithData[int]()
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	factory.ExpectNew(expectedCtx, expectedRequest, nil)
	client.ExpectDo(expectedRequest, expectedResponse, nil)
	converter.ExpectConvert(expectedCtx, expectedResponse, 0, expectedErr)

	actual, actualErr := r.RunCtx(expectedCtx, task.DoCtx)
	suite.Zero(actual)
	suite.Same(expectedErr, actualErr)

	factory.AssertExpectations(suite.T())
	client.AssertExpectations(suite.T())
	converter.AssertExpectations(suite.T())
}

func (suite *TaskSuite) testDoCtxSuccess() {
	var (
		factory   = new(mockRequestFactory)
		client    = new(mockClient)
		converter = new(mockConverter[int])

		expectedCtx      = suite.expectedCtx()
		expectedRequest  = httptest.NewRequest("GET", "/", nil)
		expectedResponse = &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(new(bytes.Buffer)),
		}

		task = Task[int]{
			Factory:   factory.New,
			Client:    client.Do,
			Converter: converter.Convert,
		}
	)

	r, err := retry.NewRunnerWithData[int]()
	suite.Require().NoError(err)
	suite.Require().NotNil(r)

	factory.ExpectNew(expectedCtx, expectedRequest, nil)
	client.ExpectDo(expectedRequest, expectedResponse, nil)
	converter.ExpectConvert(expectedCtx, expectedResponse, 123, nil)

	actual, actualErr := r.RunCtx(expectedCtx, task.DoCtx)
	suite.Equal(123, actual)
	suite.NoError(actualErr)

	factory.AssertExpectations(suite.T())
	client.AssertExpectations(suite.T())
	converter.AssertExpectations(suite.T())
}

func (suite *TaskSuite) TestDoCtx() {
	suite.Run("NoFactory", suite.testDoCtxNoFactory)
	suite.Run("FactoryError", suite.testDoCtxFactoryError)
	suite.Run("ClientError", suite.testDoCtxClientError)
	suite.Run("ConverterError", suite.testDoCtxConverterError)
	suite.Run("Success", suite.testDoCtxSuccess)
}

func (suite *TaskSuite) TestClient() {
	// verify that with no Client set, a default client is used
	suite.NotNil(Task[bool]{}.client())
}

func TestTask(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}
