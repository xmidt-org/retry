// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockTimer struct {
	mock.Mock
}

func (m *mockTimer) Timer(d time.Duration) (<-chan time.Time, func() bool) {
	args := m.Called(d)
	ch, _ := args.Get(0).(<-chan time.Time)
	stop, _ := args.Get(1).(func() bool)
	return ch, stop
}

func (m *mockTimer) Stop() bool {
	return m.Called().Bool(0)
}

func (m *mockTimer) ExpectStop(result bool) *mock.Call {
	return m.On("Stop").Return(result)
}

// ExpectTimer sets up a Timer expectation using Stop as the stop function.
func (m *mockTimer) ExpectTimer(d time.Duration, ch <-chan time.Time) *mock.Call {
	return m.On("Timer", d).Return(ch, m.Stop)
}

// ExpectConstant sets up a Timer expectation, using Stop as the stop function.
// A timer channel is used that has the supplied number of time objects already
// on the channel.
func (m *mockTimer) ExpectConstant(d time.Duration, times int) *mock.Call {
	var (
		current = time.Now()
		ch      = make(chan time.Time, times)
	)

	for i := 0; i < times; i++ {
		ch <- current
		current = current.Add(d)
	}

	return m.ExpectTimer(d, ch)
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

type mockOnAttempt struct {
	mock.Mock
}

func (m *mockOnAttempt) OnAttempt(a Attempt) {
	m.Called(a)
}

func (m *mockOnAttempt) Expect(a Attempt) *mock.Call {
	return m.On("OnAttempt", a)
}

func (m *mockOnAttempt) ExpectMatch(mf func(Attempt) bool) *mock.Call {
	return m.On("OnAttempt", mock.MatchedBy(mf))
}

type mockTask[V any] struct {
	mock.Mock
}

func (m *mockTask[V]) Do(ctx context.Context) (V, error) {
	args := m.Called(ctx)
	return args.Get(0).(V), args.Error(1)
}

func (m *mockTask[V]) Expect(ctx context.Context, result V, err error) *mock.Call {
	return m.On("Do", ctx).Return(result, err)
}

func (m *mockTask[V]) ExpectMatch(mf func(context.Context) bool, result V, err error) *mock.Call {
	return m.On("Do", mock.MatchedBy(mf)).Return(result, err)
}
