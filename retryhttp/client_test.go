package retryhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/retry"
)

var testResponse = []byte("test response")

type testHandler struct {
	lock             sync.Mutex
	expectedHeader   http.Header
	expectedAttempts int
	attempts         int
}

func (th *testHandler) resetAttempts(expectedAttempts int, expectedHeader http.Header) {
	defer th.lock.Unlock()
	th.lock.Lock()

	th.expectedHeader = expectedHeader
	th.expectedAttempts = expectedAttempts
	th.attempts = 0
}

func (th *testHandler) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	defer th.lock.Unlock()
	th.lock.Lock()

	for name := range th.expectedHeader {
		if !slices.Equal(th.expectedHeader[name], request.Header[name]) {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	th.attempts++
	if th.attempts >= th.expectedAttempts {
		rw.WriteHeader(http.StatusOK)
	} else {
		rw.WriteHeader(http.StatusServiceUnavailable)
	}

	rw.Write(testResponse)
}

type ClientSuite struct {
	suite.Suite
	th testHandler

	server *httptest.Server
}

func (suite *ClientSuite) SetupSuite() {
	suite.server = httptest.NewServer(&suite.th)
}

func (suite *ClientSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *ClientSuite) newRunner(opts ...retry.RunnerOption[*http.Response]) retry.Runner[*http.Response] {
	runner, err := retry.NewRunner(opts...)
	suite.Require().NoError(err)
	suite.Require().NotNil(runner)
	return runner
}

func (suite *ClientSuite) newClient(opts ...ClientOption) *Client {
	c, err := NewClient(opts...)
	suite.Require().NoError(err)
	suite.Require().NotNil(c)
	return c
}

func (suite *ClientSuite) newTestGet() *http.Request {
	request, err := http.NewRequest("GET", suite.server.URL+"/test", nil)
	suite.Require().NoError(err)
	suite.Require().NotNil(request)
	return request
}

func (suite *ClientSuite) assertSuccess(response *http.Response, err error) {
	suite.NoError(err)
	suite.Require().NotNil(response)
	suite.Equal(http.StatusOK, response.StatusCode)

	var o bytes.Buffer
	io.Copy(&o, response.Body)
	response.Body.Close()
	suite.Equal(testResponse, o.Bytes())
}

func (suite *ClientSuite) testGetDefault() {
	c := suite.newClient()
	suite.th.resetAttempts(1, nil)
	suite.assertSuccess(c.Do(suite.newTestGet()))
}

func (suite *ClientSuite) testGetDefaultWithRequesters() {
	c := suite.newClient(
		WithRequesters(
			func(request *http.Request) *http.Request {
				request.Header.Set("Test1", "true")
				return request
			},
			func(request *http.Request) *http.Request {
				request.Header.Set("Test2", "true")
				return request
			},
		),
	)
	suite.th.resetAttempts(1, http.Header{"Test1": []string{"true"}, "Test2": []string{"true"}})
	suite.assertSuccess(c.Do(suite.newTestGet()))
}

func (suite *ClientSuite) testGet(c *Client, expectedHeader http.Header) {
	suite.th.resetAttempts(3, expectedHeader)
	suite.assertSuccess(c.Do(suite.newTestGet()))
}

func (suite *ClientSuite) TestGet() {
	suite.Run("Default", suite.testGetDefault)
	suite.Run("DefaultWithRequesters", suite.testGetDefaultWithRequesters)

	suite.Run("WithRequesters", func() {
		c := suite.newClient(
			WithRunner(
				suite.newRunner(
					retry.WithPolicyFactory[*http.Response](
						retry.Config{
							Interval: 5 * time.Second, // won't matter due to the immediate timer
						},
					),
					retry.WithOnAttempt(CleanupResponse),
					WithShouldRetry(http.StatusServiceUnavailable),
					retry.WithImmediateTimer[*http.Response](),
				),
			),
			WithRequesters(
				func(request *http.Request) *http.Request {
					request.Header.Set("Test", "true")
					return request
				},
			),
		)

		suite.testGet(c, http.Header{"Test": []string{"true"}})
	})

	suite.Run("WithCustomClient", func() {
		c := suite.newClient(
			WithHTTPClient(new(http.Client)),
			WithRunner(
				suite.newRunner(
					retry.WithPolicyFactory[*http.Response](
						retry.Config{
							Interval: 5 * time.Second, // won't matter due to the immediate timer
						},
					),
					retry.WithOnAttempt(CleanupResponse),
					WithShouldRetry(http.StatusServiceUnavailable),
					retry.WithImmediateTimer[*http.Response](),
				),
			),
			WithRequesters(
				func(request *http.Request) *http.Request {
					request.Header.Set("Test", "true")
					return request
				},
			),
		)

		suite.testGet(c, http.Header{"Test": []string{"true"}})
	})
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
