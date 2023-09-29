// SPDX-FileCopyrightText: 2023 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package retryhttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConvertersSuite struct {
	suite.Suite
}

func (suite *ConvertersSuite) TestBoolConverter() {
	testCases := []struct {
		statusCode int
		expected   bool
	}{
		{
			statusCode: 100,
			expected:   false,
		},
		{
			statusCode: 200,
			expected:   true,
		},
		{
			statusCode: 202,
			expected:   true,
		},
		{
			statusCode: 307,
			expected:   false,
		},
		{
			statusCode: 404,
			expected:   false,
		},
		{
			statusCode: 500,
			expected:   false,
		},
	}

	for _, testCase := range testCases {
		suite.Run(fmt.Sprintf("statusCode=%d", testCase.statusCode), func() {
			actual, err := BoolConverter(
				context.Background(),
				&http.Response{
					StatusCode: testCase.statusCode,
				},
			)

			suite.NoError(err)
			suite.Equal(testCase.expected, actual)
		})
	}
}

func (suite *ConvertersSuite) TestByteConverter() {
	suite.Run("Empty", func() {
		raw, err := ByteConverter(
			context.Background(),
			&http.Response{
				Body: io.NopCloser(new(bytes.Buffer)),
			},
		)

		suite.NoError(err)
		suite.Empty(raw)
	})

	suite.Run("NotEmpty", func() {
		raw, err := ByteConverter(
			context.Background(),
			&http.Response{
				Body: io.NopCloser(bytes.NewBufferString("not empty")),
			},
		)

		suite.NoError(err)
		suite.Equal("not empty", string(raw))
	})
}

func (suite *ConvertersSuite) TestStringConverter() {
	suite.Run("Empty", func() {
		raw, err := StringConverter(
			context.Background(),
			&http.Response{
				Body: io.NopCloser(new(bytes.Buffer)),
			},
		)

		suite.NoError(err)
		suite.Empty(raw)
	})

	suite.Run("NotEmpty", func() {
		raw, err := StringConverter(
			context.Background(),
			&http.Response{
				Body: io.NopCloser(bytes.NewBufferString("not empty")),
			},
		)

		suite.NoError(err)
		suite.Equal("not empty", raw)
	})
}

func (suite *ConvertersSuite) TestJsonConverter() {
	type Custom struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	actual, err := JsonConverter[Custom](
		context.Background(),
		&http.Response{
			Body: io.NopCloser(
				bytes.NewBufferString(
					`{"name": "Joe Schmoe", "age": 92}`,
				),
			),
		},
	)

	suite.NoError(err)
	suite.Equal(
		Custom{
			Name: "Joe Schmoe",
			Age:  92,
		},
		actual,
	)
}

func TestConverters(t *testing.T) {
	suite.Run(t, new(ConvertersSuite))
}
