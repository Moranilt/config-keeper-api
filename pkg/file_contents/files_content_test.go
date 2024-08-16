package file_contents

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	database_mock "github.com/Moranilt/http-utils/clients/database/mock"
	"github.com/Moranilt/http-utils/query"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/stretchr/testify/assert"
)

func TestClient_CreateFileContent(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *CreateRequest
		mockSetup      func()
		expectedResult *FileContent
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &CreateRequest{
				FileID:   "file_id",
				Version:  "v1.0.0",
				Content:  "content",
				FormatID: "format_id",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GET_FILES_CONTENT_ID_BY_VERSION)).WithArgs("file_id", "v1.0.0").WillReturnRows(
					sqlMock.NewRows([]string{"id"}),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_CONTENT)).WithArgs("file_id", "v1.0.0", utils.StringToBase64("content"), "format_id").WillReturnRows(
					sqlMock.NewRows([]string{"id", "file_id", "version", "format", "content", "created_at", "updated_at"}).
						AddRow("file_content_id", "file_id", "v1.0.0", "yaml", utils.StringToBase64("content"), "file_content_created_at", "file_content_updated_at"),
				)
			},
			expectedResult: &FileContent{
				ID:        "file_content_id",
				FileID:    "file_id",
				Version:   "v1.0.0",
				Content:   utils.StringToBase64("content"),
				Format:    "yaml",
				CreatedAt: "file_content_created_at",
				UpdatedAt: "file_content_updated_at",
			},
			expectedError: nil,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "sql error",
			req: &CreateRequest{
				FileID:   "file_id",
				Version:  "v1.0.0",
				Content:  "content",
				FormatID: "format_id",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GET_FILES_CONTENT_ID_BY_VERSION)).WithArgs("file_id", "v1.0.0").WillReturnRows(
					sqlMock.NewRows([]string{"id"}),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_CONTENT)).WithArgs("file_id", "v1.0.0", utils.StringToBase64("content"), "format_id").WillReturnError(
					assert.AnError,
				)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "content already exists",
			req: &CreateRequest{
				FileID:   "file_id",
				Version:  "v1.0.0",
				Content:  "content",
				FormatID: "format_id",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GET_FILES_CONTENT_ID_BY_VERSION)).WithArgs("file_id", "v1.0.0").WillReturnRows(
					sqlMock.NewRows([]string{"id"}).AddRow("file_content_id"),
				)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file content already exists")),
		},
		{
			name: "empty fields in request",
			req: &CreateRequest{
				FileID:   "",
				Version:  "",
				Content:  "",
				FormatID: "",
			},
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("file_id", "required"),
				tiny_errors.Detail("version", "required"),
				tiny_errors.Detail("content", "required"),
				tiny_errors.Detail("format_id", "required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			fileContent, err := client.Create(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err, "error")
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode(), "error code")
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage(), "error message")
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails(), "error details")
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, fileContent)
		})
	}
}

func TestClient_GetFileContents(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name             string
		req              *GetManyRequest
		expectedContents []*FileContent
		expectedError    tiny_errors.ErrorHandler
		mockSetup        func()
	}{
		{
			name: "success without version",
			req: &GetManyRequest{
				FileID: "file_id",
			},
			expectedContents: []*FileContent{
				{
					ID:        "file_content_id_1",
					FileID:    "file_id_1",
					Version:   "v1.0.0",
					Content:   "content_1",
					CreatedAt: "file_content_created_at_1",
					UpdatedAt: "file_content_updated_at_1",
				},
				{
					ID:        "file_content_id_2",
					FileID:    "file_id_2",
					Version:   "v1.0.1",
					Content:   "content_2",
					CreatedAt: "file_content_created_at_2",
					UpdatedAt: "file_content_updated_at_2",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILE_CONTENTS)
				preparedQuery.Where().EQ("file_id", "file_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id", "file_id", "version", "content", "created_at", "updated_at"}).
						AddRow("file_content_id_1", "file_id_1", "v1.0.0", "content_1", "file_content_created_at_1", "file_content_updated_at_1").
						AddRow("file_content_id_2", "file_id_2", "v1.0.1", "content_2", "file_content_created_at_2", "file_content_updated_at_2"),
				)
			},
		},
		{
			name: "success with version",
			req: &GetManyRequest{
				FileID:  "file_id",
				Version: utils.MakePointer("v1.0.0"),
			},
			expectedContents: []*FileContent{
				{
					ID:        "file_content_id_1",
					FileID:    "file_id_1",
					Version:   "v1.0.0",
					Content:   "content_1",
					CreatedAt: "file_content_created_at_1",
					UpdatedAt: "file_content_updated_at_1",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILE_CONTENTS)
				preparedQuery.Where().EQ("file_id", "file_id").EQ("version", "v1.0.0")
				sqlMock.ExpectQuery(preparedQuery.String()).WillReturnRows(
					sqlMock.NewRows([]string{"id", "file_id", "version", "content", "created_at", "updated_at"}).
						AddRow("file_content_id_1", "file_id_1", "v1.0.0", "content_1", "file_content_created_at_1", "file_content_updated_at_1"),
				)
			},
		},
		{
			name: "sql error",
			req: &GetManyRequest{
				FileID: "file_id",
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("sql error")),
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILE_CONTENTS)
				preparedQuery.Where().EQ("file_id", "file_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnError(errors.New("sql error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			fileContents, err := client.GetMany(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, fileContents)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContents, fileContents)
			}
		})
	}
}

func TestClient_EditFileContents(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})
	base64Content := utils.StringToBase64("content")

	tests := []struct {
		name            string
		req             *EditRequest
		mockSetup       func()
		expectedContent *FileContent
		expectedError   tiny_errors.ErrorHandler
	}{
		{
			name: "success with version and content",
			req: &EditRequest{
				FileContentID: "file_content_id",
				Content:       utils.MakePointer("content"),
				Version:       utils.MakePointer("v1.0.0"),
			},
			mockSetup: func() {
				contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID)
				contentQuery.Where().EQ("id", "file_content_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(contentQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}).AddRow("file_content_id"),
				)

				preparedQueryUpdate := fmt.Sprintf("UPDATE file_contents SET updated_at = now(), version = 'v1.0.0', content = '%s' WHERE id = 'file_content_id' RETURNING id, file_id, version, content, created_at, updated_at", base64Content)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQueryUpdate)).WillReturnRows(
					sqlMock.NewRows([]string{"id", "file_id", "version", "content", "created_at", "updated_at"}).AddRow(
						"file_content_id", "file_id", "v1.0.0", base64Content, "file_content_created_at", "file_content_updated_at",
					),
				)
			},
			expectedContent: &FileContent{
				ID:        "file_content_id",
				FileID:    "file_id",
				Version:   "v1.0.0",
				Content:   base64Content,
				CreatedAt: "file_content_created_at",
				UpdatedAt: "file_content_updated_at",
			},
			expectedError: nil,
		},
		{
			name: "success with version",
			req: &EditRequest{
				FileContentID: "file_content_id",
				Version:       utils.MakePointer("v1.0.0"),
			},
			mockSetup: func() {
				contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID)
				contentQuery.Where().EQ("id", "file_content_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(contentQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}).AddRow("file_content_id"),
				)

				preparedQueryUpdate := "UPDATE file_contents SET updated_at = now(), version = 'v1.0.0' WHERE id = 'file_content_id' RETURNING id, file_id, version, content, created_at, updated_at"
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQueryUpdate)).WillReturnRows(
					sqlMock.NewRows([]string{"id", "file_id", "version", "content", "created_at", "updated_at"}).AddRow(
						"file_content_id", "file_id", "v1.0.0", base64Content, "file_content_created_at", "file_content_updated_at",
					),
				)
			},
			expectedContent: &FileContent{
				ID:        "file_content_id",
				FileID:    "file_id",
				Version:   "v1.0.0",
				Content:   base64Content,
				CreatedAt: "file_content_created_at",
				UpdatedAt: "file_content_updated_at",
			},
			expectedError: nil,
		},
		{
			name: "success with content",
			req: &EditRequest{
				FileContentID: "file_content_id",
				Content:       utils.MakePointer("content"),
			},
			mockSetup: func() {
				contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID)
				contentQuery.Where().EQ("id", "file_content_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(contentQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}).AddRow("file_content_id"),
				)

				preparedQueryUpdate := fmt.Sprintf("UPDATE file_contents SET updated_at = now(), content = '%s' WHERE id = 'file_content_id' RETURNING id, file_id, version, content, created_at, updated_at", base64Content)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQueryUpdate)).WillReturnRows(
					sqlMock.NewRows([]string{"id", "file_id", "version", "content", "created_at", "updated_at"}).AddRow(
						"file_content_id", "file_id", "v1.0.0", base64Content, "file_content_created_at", "file_content_updated_at",
					),
				)
			},
			expectedContent: &FileContent{
				ID:        "file_content_id",
				FileID:    "file_id",
				Version:   "v1.0.0",
				Content:   base64Content,
				CreatedAt: "file_content_created_at",
				UpdatedAt: "file_content_updated_at",
			},
			expectedError: nil,
		},
		{
			name: "failed without version and content",
			req: &EditRequest{
				FileContentID: "file_content_id",
			},
			mockSetup:       func() {},
			expectedContent: nil,
			expectedError:   tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD),
		},
		{
			name: "failed content query",
			req: &EditRequest{
				FileContentID: "file_content_id",
				Content:       utils.MakePointer("content"),
			},
			mockSetup: func() {
				contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID)
				contentQuery.Where().EQ("id", "file_content_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(contentQuery.String())).WillReturnError(errors.New("sql error"))
			},
			expectedContent: nil,
			expectedError:   tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("sql error")),
		},
		{
			name: "update error",
			req: &EditRequest{
				FileContentID: "file_content_id",
				Content:       utils.MakePointer("content"),
			},
			mockSetup: func() {
				contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID)
				contentQuery.Where().EQ("id", "file_content_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(contentQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}).AddRow("file_content_id"),
				)

				preparedQueryUpdate := fmt.Sprintf("UPDATE file_contents SET updated_at = now(), content = '%s' WHERE id = 'file_content_id' RETURNING id, file_id, version, content, created_at, updated_at", base64Content)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQueryUpdate)).WillReturnError(errors.New("sql error"))
			},
			expectedContent: nil,
			expectedError:   tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("sql error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			fileContent, err := client.Edit(context.Background(), tt.req)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, fileContent)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, fileContent)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name            string
		req             *DeleteRequest
		mockSetup       func()
		expectedRemoved bool
		expectedError   tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &DeleteRequest{
				ID: "123",
			},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FILE_CONTENT)).WithArgs("123").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedRemoved: true,
			expectedError:   nil,
		},
		{
			name: "not found element, rows affected 0",
			req: &DeleteRequest{
				ID: "123",
			},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FILE_CONTENT)).WithArgs("123").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedRemoved: false,
			expectedError:   nil,
		},
		{
			name:            "empty request",
			req:             nil,
			mockSetup:       func() {},
			expectedRemoved: false,
			expectedError:   tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &DeleteRequest{
				ID: "123",
			},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FILE_CONTENT)).WithArgs("123").WillReturnError(assert.AnError)
			},
			expectedRemoved: false,
			expectedError:   tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			removed, err := client.Delete(context.Background(), tt.req)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedRemoved, removed)
		})
	}
}
