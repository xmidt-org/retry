package retry

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockSleep struct {
	mock.Mock
}

func (m *mockSleep) Sleep(d time.Duration) {
	m.Called(d)
}

func (m *mockSleep) Expect(d time.Duration) *mock.Call {
	return m.On("Sleep", d)
}

type mockShouldRetry struct {
	mock.Mock
}

func (m *mockShouldRetry) ShouldRetry(err error) bool {
	return m.Called(err).Bool(0)
}

func (m *mockShouldRetry) Expect(err error, result bool) *mock.Call {
	return m.On("ShouldRetry", err).Return(result)
}

type mockOnFail struct {
	mock.Mock
}

func (m *mockOnFail) OnFail(err error, d time.Duration) {
	m.Called(err, d)
}

func (m *mockOnFail) Expect(err error, d time.Duration) *mock.Call {
	return m.On("OnFail", err, d)
}

type mockTask struct {
	mock.Mock
}

func (m *mockTask) Do() error {
	return m.Called().Error(0)
}

func (m *mockTask) Expect(err error) *mock.Call {
	return m.On("Do").Return(err)
}

func (m *mockTask) DoCtx(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockTask) ExpectCtx(ctx context.Context, err error) *mock.Call {
	return m.On("DoCtx", ctx).Return(err)
}

type mockTaskWithData[V any] struct {
	mock.Mock
}

func (m *mockTaskWithData[V]) Do() (V, error) {
	args := m.Called()
	return args.Get(0).(V), args.Error(1)
}

func (m *mockTaskWithData[V]) Expect(result V, err error) *mock.Call {
	return m.On("Do").Return(result, err)
}

func (m *mockTaskWithData[V]) DoCtx(ctx context.Context) (V, error) {
	args := m.Called(ctx)
	return args.Get(0).(V), args.Error(1)
}

func (m *mockTaskWithData[V]) ExpectCtx(ctx context.Context, result V, err error) *mock.Call {
	return m.On("DoCtx", ctx).Return(result, err)
}
