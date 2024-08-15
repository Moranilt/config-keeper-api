package callback

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/pkg/file_contents"
	"github.com/Moranilt/config-keeper/pkg/files"
	"github.com/Moranilt/config-keeper/pkg/listeners"
	"github.com/Moranilt/http-utils/logger"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPrepareListenersData(t *testing.T) {
	// Create mock services
	mockFile := files.NewMock()
	mockContent := file_contents.NewMock()
	mockListeners := listeners.NewMock()
	mockLog := logger.NewMock()
	sendChannel := NewChannel(1)
	// Create service with mocks
	service := New(mockLog, sendChannel, mockFile, mockListeners, mockContent, nil)

	setupMocks := func(
		fileID string,
		file *files.File,
		fileError tiny_errors.ErrorHandler,
		fileContents []*file_contents.FileContent,
		fileContentsError tiny_errors.ErrorHandler,
		listenersList []*listeners.Listener,
		listenersError tiny_errors.ErrorHandler,
	) {
		if fileError != nil {
			mockFile.On("Get", mock.Anything, &files.GetRequest{ID: fileID}).Return(nil, fileError)
		} else {
			mockFile.On("Get", mock.Anything, &files.GetRequest{ID: fileID}).Return(file, nil)
		}
		if fileContentsError != nil {
			mockContent.On("GetMany", mock.Anything, &file_contents.GetManyRequest{FileID: fileID}).Return(nil, fileContentsError)
		} else if fileError == nil {
			mockContent.On("GetMany", mock.Anything, &file_contents.GetManyRequest{FileID: fileID}).Return(fileContents, nil)
		}

		if listenersError != nil {
			mockListeners.On("GetMany", mock.Anything, &listeners.GetManyRequest{FileID: fileID}).Return(nil, listenersError)
		} else if fileError == nil && fileContentsError == nil {
			mockListeners.On("GetMany", mock.Anything, &listeners.GetManyRequest{FileID: fileID}).Return(listenersList, nil)
		}
	}

	tests := []struct {
		name              string
		req               *CallbackRequest
		file              *files.File
		fileError         tiny_errors.ErrorHandler
		fileContents      []*file_contents.FileContent
		fileContentsError tiny_errors.ErrorHandler
		listeners         []*listeners.Listener
		listenersError    tiny_errors.ErrorHandler
		getExpectedData   func(*files.File, []*file_contents.FileContent) []byte
		expectedError     bool
	}{
		{
			name: "Successful preparation",
			req:  &CallbackRequest{FileID: "file1"},
			file: &files.File{ID: "file1", Name: "test.txt"},
			fileContents: []*file_contents.FileContent{
				{FileID: "file1", Content: "test content"},
			},
			listeners: []*listeners.Listener{
				{FileID: "file1", CallbackEndpoint: "http://example.com/callback1"},
			},
			getExpectedData: func(file *files.File, fileContents []*file_contents.FileContent) []byte {
				fileData := &FileData{
					File:        *file,
					FileContent: fileContents,
				}
				data, _ := json.Marshal(fileData)
				return data
			},
			expectedError: false,
		},
		{
			name:          "Error getting file",
			req:           &CallbackRequest{FileID: "nonexistent"},
			fileError:     tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("file not found")),
			expectedError: true,
		},
		{
			name:              "Error getting file contents",
			req:               &CallbackRequest{FileID: "file2"},
			file:              &files.File{ID: "file2", Name: "error.txt"},
			fileContentsError: tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("content not found")),
			expectedError:     true,
		},
		{
			name: "Error getting listeners",
			req:  &CallbackRequest{FileID: "file3"},
			file: &files.File{ID: "file3", Name: "listeners.txt"},
			fileContents: []*file_contents.FileContent{
				{FileID: "file3", Content: "content"},
			},
			listenersError: tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("listeners not found")),
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations
			setupMocks(tt.req.FileID, tt.file, tt.fileError, tt.fileContents, tt.fileContentsError, tt.listeners, tt.listenersError)

			// Call the function
			listeners, data, err := service.prepareListenersData(context.Background(), tt.req)

			// Assert expectations
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, listeners)
				assert.Nil(t, data)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, listeners)
				assert.Equal(t, tt.getExpectedData(tt.file, tt.fileContents), data)
			}

			mockFile.AssertExpectations(t)
			mockContent.AssertExpectations(t)
			mockListeners.AssertExpectations(t)
		})
	}
}
