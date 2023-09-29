// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) Do(request *http.Request) (*http.Response, error) {
	args := m.Called(request)
	response, _ := args.Get(0).(*http.Response)
	return response, args.Error(1)
}

func (m *mockClient) ExpectDo(request *http.Request, response *http.Response, err error) *mock.Call {
	return m.On("Do", request).Return(response, err)
}

type mockRequestFactory struct {
	mock.Mock
}

func (m *mockRequestFactory) New(ctx context.Context) (*http.Request, error) {
	args := m.Called(ctx)
	request, _ := args.Get(0).(*http.Request)
	return request, args.Error(1)
}

func (m *mockRequestFactory) ExpectNew(ctx context.Context, request *http.Request, err error) *mock.Call {
	return m.On("New", ctx).Return(request, err)
}

type mockConverter[V any] struct {
	mock.Mock
}

func (m *mockConverter[V]) Convert(ctx context.Context, response *http.Response) (V, error) {
	args := m.Called(ctx, response)
	result, _ := args.Get(0).(V)
	return result, args.Error(1)
}

func (m *mockConverter[V]) ExpectConvert(ctx context.Context, response *http.Response, result V, err error) *mock.Call {
	return m.On("Convert", ctx, response).Return(result, err)
}
