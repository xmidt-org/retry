// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func ExampleRunner_never() {
	onAttempt := func(a Attempt[int]) {
		fmt.Println("Attempt.Result", a.Result, "Attempt.Next", a.Next)
	}

	// no Config or PolicyFactory means this runner will never
	// retry anything
	runner, _ := NewRunner[int](
		WithOnAttempt(onAttempt),
	)

	result, _ := runner.Run(
		context.Background(),
		AddValue(1234, func() error {
			fmt.Println("executing task ...")
			return nil
		}),
	)

	fmt.Println("Run result", result)

	// Output:
	// executing task ...
	// Attempt.Result 1234 Attempt.Next 0s
	// Run result 1234
}

func ExampleRunner_constant() {
	onAttempt := func(a Attempt[int]) {
		fmt.Println("Attempt.Result", a.Result, "Attempt.Next", a.Next)
	}

	runner, _ := NewRunner(
		WithOnAttempt(onAttempt),
		WithPolicyFactory[int](Config{
			Interval: 10 * time.Millisecond,
		}),
	)

	attempts := 0

	result, _ := runner.Run(
		context.Background(),
		func(_ context.Context) (int, error) {
			fmt.Println("executing task ...")
			attempts++
			if attempts < 3 {
				return -1, errors.New("task error")
			}

			return 1234, nil
		},
	)

	fmt.Println("Run result", result)

	// Output:
	// executing task ...
	// Attempt.Result -1 Attempt.Next 10ms
	// executing task ...
	// Attempt.Result -1 Attempt.Next 10ms
	// executing task ...
	// Attempt.Result 1234 Attempt.Next 0s
	// Run result 1234
}
