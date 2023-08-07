package retryhttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/xmidt-org/retry"
)

func ExampleTask_DoCtx() {
	r, err := retry.NewRunnerWithData[bool](
		retry.WithPolicyFactory(retry.Config{
			// desired configuration ...
		}),
	)

	if err != nil {
		panic(err)
	}

	task := Task[bool]{
		Factory: func(ctx context.Context) (*http.Request, error) {
			fmt.Println("creating request")
			return http.NewRequestWithContext(ctx, "GET", "/", nil)
		},
		Client: func(*http.Request) (*http.Response, error) {
			fmt.Println("executing HTTP transaction")
			return &http.Response{
				StatusCode: 200,
			}, nil
		},
		Converter: func(ctx context.Context, response *http.Response) (bool, error) {
			fmt.Println("converting response")

			// we normally would use an error for a non-2xx status code, but this is just an example
			return response.StatusCode == http.StatusOK, nil
		},
	}

	result, taskErr := r.RunCtx(context.Background(), task.DoCtx)
	if taskErr != nil {
		panic(taskErr)
	}

	fmt.Println("result", result)

	// Output:
	// creating request
	// executing HTTP transaction
	// converting response
	// result true
}
