package listeners

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

func (m *MockClient) Create(ctx context.Context, req *CreateRequest) (*Listener, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	listener := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return listener.(*Listener), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) GetMany(ctx context.Context, req *GetManyRequest) ([]*Listener, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	listeners := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return listeners.([]*Listener), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}

func (m *MockClient) Get(ctx context.Context, req *GetRequest) (*Listener, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	listener := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return listener.(*Listener), nil
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

func (m *MockClient) Edit(ctx context.Context, req *EditRequest) (*Listener, tiny_errors.ErrorHandler) {
	args := m.Called(ctx, req)
	listener := args.Get(0)
	err := args.Get(1)
	if err == nil {
		return listener.(*Listener), nil
	}
	return nil, err.(tiny_errors.ErrorHandler)
}
