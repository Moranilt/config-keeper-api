package listeners

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	database_mock "github.com/Moranilt/http-utils/clients/database/mock"
	"github.com/Moranilt/http-utils/query"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/stretchr/testify/assert"
)

func TestClientCreate(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *CreateRequest
		mockSetup      func()
		expectedResult *Listener
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "Create listener successfully",
			req: &CreateRequest{
				FileID:           "file123",
				CallbackEndpoint: "http://example.com/callback",
				Name:             "Test Listener",
			},
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "file_id", "callback_endpoint", "name", "created_at", "updated_at"}).
					AddRow("listener123", "file123", "http://example.com/callback", "Test Listener", time.Now(), time.Now())
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_LISTENER)).
					WithArgs("file123", "http://example.com/callback", "Test Listener").
					WillReturnRows(rows)
			},
			expectedResult: &Listener{
				ID:               "listener123",
				FileID:           "file123",
				CallbackEndpoint: "http://example.com/callback",
				Name:             "Test Listener",
			},
		},
		{
			name:          "Create listener with nil request",
			req:           nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "Database error on insert",
			req: &CreateRequest{
				FileID:           "file456",
				CallbackEndpoint: "http://example.com/callback2",
				Name:             "Test Listener 2",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_LISTENER)).
					WithArgs("file456", "http://example.com/callback2", "Test Listener 2").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database),
		},
		{
			name: "missing file_id",
			req: &CreateRequest{
				CallbackEndpoint: "http://example.com/callback",
				Name:             "Test Listener",
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("file_id", "required")),
		},
		{
			name: "missing callback_endpoint",
			req: &CreateRequest{
				FileID: "file123",
				Name:   "Test Listener",
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("callback_endpoint", "required")),
		},
		{
			name: "missing name",
			req: &CreateRequest{
				FileID:           "file123",
				CallbackEndpoint: "http://example.com/callback",
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("name", "required")),
		},
		{
			name: "all fields missing",
			req:  &CreateRequest{},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("file_id", "required"),
				tiny_errors.Detail("callback_endpoint", "required"),
				tiny_errors.Detail("name", "required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			listener, err := client.Create(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, listener)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				if len(tt.expectedError.GetDetails()) > 0 {
					assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, listener)
				assert.Equal(t, tt.expectedResult.ID, listener.ID)
				assert.Equal(t, tt.expectedResult.FileID, listener.FileID)
				assert.Equal(t, tt.expectedResult.CallbackEndpoint, listener.CallbackEndpoint)
				assert.Equal(t, tt.expectedResult.Name, listener.Name)
			}
		})
	}
}

func TestClient_GetMany(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name              string
		req               *GetManyRequest
		mockSetup         func()
		expectedListeners []*Listener
		expectedError     tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &GetManyRequest{
				FileID: "file123",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_LISTENERS)
				preparedQuery.Where().EQ("file_id", "file123")
				preparedQuery.Order("name", "asc")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlmock.NewRows([]string{"id", "file_id", "name", "callback_endpoint"}).
						AddRow("listener1", "file123", "Listener 1", "http://example.com/1").
						AddRow("listener2", "file123", "Listener 2", "http://example.com/2"),
				)
			},
			expectedListeners: []*Listener{
				{ID: "listener1", FileID: "file123", Name: "Listener 1", CallbackEndpoint: "http://example.com/1"},
				{ID: "listener2", FileID: "file123", Name: "Listener 2", CallbackEndpoint: "http://example.com/2"},
			},
			expectedError: nil,
		},
		{
			name:              "empty request",
			req:               nil,
			mockSetup:         func() {},
			expectedListeners: nil,
			expectedError:     tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "database error",
			req: &GetManyRequest{
				FileID: "file123",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_LISTENERS)
				preparedQuery.Where().EQ("file_id", "file123")
				preparedQuery.Order("name", "asc")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnError(errors.New("database error"))
			},
			expectedListeners: nil,
			expectedError:     tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			listeners, err := client.GetMany(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, listeners)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedListeners, listeners)
			}
		})
	}
}

func TestClient_Create(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name             string
		req              *CreateRequest
		mockSetup        func()
		expectedListener *Listener
		expectedError    tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &CreateRequest{
				FileID:           "file123",
				CallbackEndpoint: "http://example.com/callback",
				Name:             "Test Listener",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_LISTENER)).
					WithArgs("file123", "http://example.com/callback", "Test Listener").
					WillReturnRows(sqlmock.NewRows([]string{"id", "file_id", "callback_endpoint", "name"}).
						AddRow("listener123", "file123", "http://example.com/callback", "Test Listener"))
			},
			expectedListener: &Listener{
				ID:               "listener123",
				FileID:           "file123",
				CallbackEndpoint: "http://example.com/callback",
				Name:             "Test Listener",
			},
			expectedError: nil,
		},
		{
			name: "missing required fields",
			req: &CreateRequest{
				FileID: "file123",
				Name:   "Test Listener",
			},
			mockSetup:        func() {},
			expectedListener: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("callback_endpoint", "required")),
		},
		{
			name: "database error",
			req: &CreateRequest{
				FileID:           "file123",
				CallbackEndpoint: "http://example.com/callback",
				Name:             "Test Listener",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_LISTENER)).
					WithArgs("file123", "http://example.com/callback", "Test Listener").
					WillReturnError(errors.New("database error"))
			},
			expectedListener: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database,
				tiny_errors.Message("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			listener, err := client.Create(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, listener)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedListener, listener)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *DeleteRequest
		mockSetup      func()
		expectedResult bool
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req:  &DeleteRequest{ID: "listener123"},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_LISTENER)).
					WithArgs("listener123").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "listener not found",
			req:  &DeleteRequest{ID: "nonexistent"},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_LISTENER)).
					WithArgs("nonexistent").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "database error",
			req:  &DeleteRequest{ID: "listener123"},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_LISTENER)).
					WithArgs("listener123").
					WillReturnError(errors.New("database error"))
			},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.Delete(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Edit(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *EditRequest
		expectedResult *Listener
		expectedError  tiny_errors.ErrorHandler
		mockSetup      func()
	}{
		{
			name: "success",
			req: &EditRequest{
				ID:               "listener_id",
				Name:             utils.MakePointer("new_name"),
				CallbackEndpoint: utils.MakePointer("new_endpoint"),
			},
			expectedResult: &Listener{
				ID:               "listener_id",
				FileID:           "file_id",
				Name:             "new_name",
				CallbackEndpoint: "new_endpoint",
				CreatedAt:        "2023-01-01T00:00:00Z",
				UpdatedAt:        "2023-01-02T00:00:00Z",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_LISTENERS)
				preparedQuery.Where().EQ("id", "listener_id")
				sqlMock.ExpectBegin()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("listener_id"))
				sqlMock.ExpectQuery("UPDATE listeners SET").
					WithArgs("new_name", "new_endpoint", "now()", "listener_id").
					WillReturnRows(sqlmock.NewRows([]string{"id", "file_id", "callback_endpoint", "name", "created_at", "updated_at"}).
						AddRow("listener_id", "file_id", "new_endpoint", "new_name", "2023-01-01T00:00:00Z", "2023-01-02T00:00:00Z"))
				sqlMock.ExpectCommit()
			},
		},
		{
			name: "listener not found",
			req: &EditRequest{
				ID:   "non_existent_id",
				Name: utils.MakePointer("new_name"),
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("listener does not exist")),
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_LISTENERS)
				preparedQuery.Where().EQ("id", "non_existent_id")
				sqlMock.ExpectBegin()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(sql.ErrNoRows)
				sqlMock.ExpectRollback()
			},
		},
		{
			name: "database error",
			req: &EditRequest{
				ID:   "listener_id",
				Name: utils.MakePointer("new_name"),
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("database error")),
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_LISTENERS)
				preparedQuery.Where().EQ("id", "listener_id")
				sqlMock.ExpectBegin()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(errors.New("database error"))
				sqlMock.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.Edit(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func TestBuildUpdateQuery(t *testing.T) {
	tests := []struct {
		name          string
		req           *EditRequest
		expectedQuery string
		expectedArgs  []interface{}
	}{
		{
			name: "update name and callback_endpoint",
			req: &EditRequest{
				ID:               "listener_id",
				Name:             utils.MakePointer("new_name"),
				CallbackEndpoint: utils.MakePointer("new_endpoint"),
			},
			expectedQuery: "UPDATE listeners SET name = $1, callback_endpoint = $2, updated_at = $3 WHERE id = $4 RETURNING id, file_id, callback_endpoint, name, created_at, updated_at",
			expectedArgs:  []interface{}{"new_name", "new_endpoint", "now()", "listener_id"},
		},
		{
			name: "update name only",
			req: &EditRequest{
				ID:   "listener_id",
				Name: utils.MakePointer("new_name"),
			},
			expectedQuery: "UPDATE listeners SET name = $1, updated_at = $2 WHERE id = $3 RETURNING id, file_id, callback_endpoint, name, created_at, updated_at",
			expectedArgs:  []interface{}{"new_name", "now()", "listener_id"},
		},
		{
			name: "update callback_endpoint only",
			req: &EditRequest{
				ID:               "listener_id",
				CallbackEndpoint: utils.MakePointer("new_endpoint"),
			},
			expectedQuery: "UPDATE listeners SET callback_endpoint = $1, updated_at = $2 WHERE id = $3 RETURNING id, file_id, callback_endpoint, name, created_at, updated_at",
			expectedArgs:  []interface{}{"new_endpoint", "now()", "listener_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := buildUpdateQuery(tt.req)
			assert.Equal(t, tt.expectedQuery, query)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}
