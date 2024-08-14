package files

import (
	"context"

	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func NewMock() *MockClient {
	return new(MockClient)
}

func (m *MockClient) GetFilesInFolder(ctx context.Context, req *GetFilesInFolderRequest) ([]*File, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	files := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return files.([]*File), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Create(ctx context.Context, req *CreateRequest) (*File, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	file := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return file.(*File), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	success := args.Bool(0)
	err := args.Get(1)
	if err == nil {
		return success, nil
	}
	return success, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Edit(ctx context.Context, req *EditRequest) (*File, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	file := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return file.(*File), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Get(ctx context.Context, req *GetRequest) (*File, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	file := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return file.(*File), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}
