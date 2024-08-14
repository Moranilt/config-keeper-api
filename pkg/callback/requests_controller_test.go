package callback

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Moranilt/http-utils/client"
	"github.com/Moranilt/http-utils/logger"
	"github.com/stretchr/testify/assert"
)

func TestSendRequestWithRetry(t *testing.T) {
	mockLog := logger.NewMock()
	mockHttpClient := client.NewMock()
	service := NewRequestsController(mockLog, mockHttpClient)

	tests := []struct {
		name           string
		endpoint       string
		data           []byte
		setupMocks     func()
		expectedError  bool
		expectedErrMsg string
		cancelContext  bool
	}{
		{
			name:     "successful request on first attempt",
			endpoint: "http://example.com",
			data:     []byte(`{"key":"value"}`),
			setupMocks: func() {
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), nil, &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(&bytes.Buffer{}),
				}, nil)
			},
			expectedError: false,
		},
		{
			name:     "successful request after one retry",
			endpoint: "http://example.com",
			data:     []byte(`{"key":"value"}`),
			setupMocks: func() {
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), nil, &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(&bytes.Buffer{}),
				}, nil)
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), nil, &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(&bytes.Buffer{}),
				}, nil)
			},
			expectedError: false,
		},
		{
			name:     "max retries reached",
			endpoint: "http://example.com",
			data:     []byte(`{"key":"value"}`),
			setupMocks: func() {
				for i := 0; i < 3; i++ {
					mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), nil, &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(&bytes.Buffer{}),
					}, nil)
				}
			},
			expectedError:  true,
			expectedErrMsg: "max retries reached for endpoint http://example.com",
		},
		{
			name:     "context cancelled",
			endpoint: "http://example.com",
			data:     []byte(`{"key":"value"}`),
			setupMocks: func() {
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), fmt.Errorf("context canceled"), nil, nil)
			},
			expectedError:  true,
			expectedErrMsg: "context canceled",
			cancelContext:  true,
		},
		{
			name:     "network error",
			endpoint: "http://example.com",
			data:     []byte(`{"key":"value"}`),
			setupMocks: func() {
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), fmt.Errorf("network error"), nil, nil)
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), fmt.Errorf("network error"), nil, nil)
				mockHttpClient.ExpectPost("http://example.com", []byte(`{"key":"value"}`), fmt.Errorf("network error"), nil, nil)
			},
			expectedError:  true,
			expectedErrMsg: "max retries reached for endpoint http://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			if tt.cancelContext {
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}
			err := service.SendRequestWithRetry(ctx, tt.endpoint, tt.data)

			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			if err := mockHttpClient.AllExpectationsDone(); err != nil {
				t.Error(err)
			}
		})
	}
}
