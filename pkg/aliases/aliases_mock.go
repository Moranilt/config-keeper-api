package aliases

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

func (m *MockClient) Create(ctx context.Context, req *CreateRequest) (*Alias, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	alias := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return alias.(*Alias), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) GetMany(ctx context.Context, req *GetManyRequest) ([]*Alias, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	aliases := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return aliases.([]*Alias), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Edit(ctx context.Context, req *EditRequest) (*Alias, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	alias := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return alias.(*Alias), nil
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

func (m *MockClient) ExistsInFile(ctx context.Context, req *ExistsInFileRequest) ([]string, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	aliases := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return aliases.([]string), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) AddToFile(ctx context.Context, req *AddToFileRequest) (int, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	aliases := args.Int(0)
	err := args.Get(1)
	if err == nil {
		return aliases, nil
	}
	return 0, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) RemoveFromFile(ctx context.Context, req *RemoveFromFileRequest) (int, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	aliases := args.Int(0)
	err := args.Get(1)
	if err == nil {
		return aliases, nil
	}
	return 0, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Get(ctx context.Context, req *GetRequest) (*Alias, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	alias := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return alias.(*Alias), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) GetFileAliases(ctx context.Context, req *GetFileAliasesRequest) ([]*Alias, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	aliases := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return aliases.([]*Alias), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}
