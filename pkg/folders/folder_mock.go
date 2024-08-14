package folders

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

func (m *MockClient) New(ctx context.Context, req *NewRequest) (*Folder, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	folder := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return folder.(*Folder), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Exists(ctx context.Context, req *ExistsRequest) (bool, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	exists := args.Bool(0)
	err := args.Get(1)
	if err == nil {
		return exists, nil
	}
	return exists, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Get(ctx context.Context, req *GetRequest) (*FolderWithPath, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	folder := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return folder.(*FolderWithPath), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) GetMany(ctx context.Context, req *GetManyRequest) ([]*Folder, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	folders := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return folders.([]*Folder), nil
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

func (m *MockClient) Edit(ctx context.Context, req *EditRequest) (*Folder, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	folder := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return folder.(*Folder), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}
