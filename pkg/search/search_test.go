package search

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	database_mock "github.com/Moranilt/http-utils/clients/database/mock"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/stretchr/testify/assert"
)

func TestGlobal(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)

	tests := []struct {
		name           string
		req            *GlobalSearchRequest
		mockSetup      func(sqlMock sqlmock.Sqlmock)
		expectedResult []*GlobalSearchResult
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "Valid request",
			req:  &GlobalSearchRequest{Query: "test"},
			mockSetup: func(sqlMock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "folder_name", "file_name", "folder_id", "aliases", "created_at", "updated_at"}).
					AddRow("3624caaa-2d98-431b-a106-f26dcb523fa3", "Test Folder", "test.txt", "b7a24e6d-5c4f-4e1e-a3c9-f264915e16c6", "test", "2022-01-25T15:04:05.000Z", "2022-01-25T15:04:05.000Z")
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GLOBAL_SEARCH)).
					WithArgs("test").
					WillReturnRows(rows)
			},
			expectedResult: []*GlobalSearchResult{
				{
					ID:         "3624caaa-2d98-431b-a106-f26dcb523fa3",
					FolderName: "Test Folder",
					FileName:   "test.txt",
					FolderID:   utils.MakePointer("b7a24e6d-5c4f-4e1e-a3c9-f264915e16c6"),
					Aliases:    "test",
					CreatedAt:  "2022-01-25T15:04:05.000Z",
					UpdatedAt:  "2022-01-25T15:04:05.000Z"}},
		},
		{
			name:          "Nil request",
			req:           nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name:          "Empty query",
			req:           &GlobalSearchRequest{Query: ""},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("query", "required")),
		},
		{
			name: "Database error",
			req:  &GlobalSearchRequest{Query: "test"},
			mockSetup: func(sqlMock sqlmock.Sqlmock) {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GLOBAL_SEARCH)).
					WithArgs("test").
					WillReturnError(assert.AnError)
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "No results",
			req:  &GlobalSearchRequest{Query: "nonexistent"},
			mockSetup: func(sqlMock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "folder_name", "file_name", "folder_id", "aliases", "created_at", "updated_at"})
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GLOBAL_SEARCH)).
					WithArgs("nonexistent").
					WillReturnRows(rows)
			},
			expectedResult: []*GlobalSearchResult{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDb, sqlMock := database_mock.NewSQlMock(t)
			client := New(&database.Client{mockDb})
			if tt.mockSetup != nil {
				tt.mockSetup(sqlMock)
			}

			result, err := client.Global(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
			assert.Nil(t, sqlMock.ExpectationsWereMet())
		})
	}
}
