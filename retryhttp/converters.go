package retryhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

// Converter converts an *http.Response into an arbitrary value. This closure is
// invoked for all non-nil responses, including non-2xx responses.  That allows
// the implementation to convert non-2xx responses into errors.
//
// Note: the original request is available via http.Response.Request.
type Converter[V any] func(context.Context, *http.Response) (V, error)

// BoolConverter is a Converter[bool] that returns true if the response is a success,
// false otherwise.
//
// For most situations, a non-2xx response should be an error with retry semantics attached.
// This converter is useful on in the simplest circumstances.
func BoolConverter(_ context.Context, response *http.Response) (bool, error) {
	return (response.StatusCode >= 200 && response.StatusCode <= 299), nil
}

// ByteConverter is a Converter[[]byte] that returns the raw response body.
func ByteConverter(_ context.Context, response *http.Response) ([]byte, error) {
	return io.ReadAll(response.Body)
}

// StringConverter is a Converter[string] that returns the response body as a UTF-8 string.
func StringConverter(_ context.Context, response *http.Response) (v string, err error) {
	var data []byte
	data, err = io.ReadAll(response.Body)
	if err == nil {
		v = string(data)
	}

	return
}

// JsonConverter uses encoding/json to unmarshal the response body.  The type V must be
// something which is unmarshalable as a *V, typically a struct or a map.
//
// This converter doesn't take into account any URL values or headers.  Only the response
// body is used to unmarshal V.
func JsonConverter[V any](_ context.Context, response *http.Response) (result V, err error) {
	var data []byte
	data, err = io.ReadAll(response.Body)
	if err == nil {
		err = json.Unmarshal(data, &result)
	}

	return
}
