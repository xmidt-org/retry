package retryhttp

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/xmidt-org/retry"
)

var testResponse = []byte("test response")

type handlerState struct {
	lock             sync.Mutex
	expectedAttempts int
	attempts         int
}

func (hs *handlerState) resetAttempts(expectedAttempts int) {
	defer hs.lock.Unlock()
	hs.lock.Lock()

	hs.expectedAttempts = expectedAttempts
	hs.attempts = 0
}

func (hs *handlerState) onAttempt() (done bool) {
	defer hs.lock.Unlock()
	hs.lock.Lock()

	hs.attempts++
	if hs.attempts >= hs.expectedAttempts {
		done = true
	}

	return
}

type ClientSuite struct {
	suite.Suite
	handlerState

	server *httptest.Server
}

func (suite *ClientSuite) SetupSuite() {
	suite.server = httptest.NewServer(suite)
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

func (suite *ClientSuite) newTestPut() *http.Request {
	request, err := http.NewRequest("PUT", suite.server.URL+"/test", strings.NewReader("test"))
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

// ServeHTTP lets this suite be an http.Handler that returns a series of http.StatusServiceUnavailable
// results followed by a final http.StatusOK.  The number of unavailable codes returned is managed
// by the handler state.
func (suite *ClientSuite) ServeHTTP(rw http.ResponseWriter, request *http.Request) {
	if suite.onAttempt() {
		rw.WriteHeader(http.StatusOK)
	} else {
		rw.WriteHeader(http.StatusServiceUnavailable)
	}

	rw.Write(testResponse)
}

func (suite *ClientSuite) TestGet() {
	suite.Run("Default", func() {
		c := suite.newClient()
		suite.resetAttempts(1)
		suite.assertSuccess(c.Do(suite.newTestGet()))
	})

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

		suite.resetAttempts(3)
		suite.assertSuccess(c.Do(suite.newTestGet()))
	})

	suite.Run("WithCustomClient", func() {
	})
}

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientSuite))
}
