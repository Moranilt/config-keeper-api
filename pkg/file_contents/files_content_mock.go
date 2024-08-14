package file_contents

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

func (m *MockClient) Create(ctx context.Context, req *CreateRequest) (*FileContent, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	content := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return content.(*FileContent), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) GetMany(ctx context.Context, req *GetManyRequest) ([]*FileContent, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	contents := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return contents.([]*FileContent), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Edit(ctx context.Context, req *EditRequest) (*FileContent, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	content := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return content.(*FileContent), nil
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
