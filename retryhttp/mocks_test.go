// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(request *http.Request) (*http.Response, error) {
	args := m.Called(request)
	response, _ := args.Get(0).(*http.Response)
	return response, args.Error(1)
}

func (m *mockHTTPClient) ExpectDo(request *http.Request, response *http.Response, err error) *mock.Call {
	return m.On("Do", request).Return(response, err)
}
