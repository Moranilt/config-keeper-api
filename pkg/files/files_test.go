package files

import (
	"context"
	"database/sql"
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

func TestClient_GetFilesInFolder(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name          string
		req           *GetFilesInFolderRequest
		expectedFiles []*File
		expectedError tiny_errors.ErrorHandler
		mockSetup     func()
	}{
		{
			name: "success with empty folder_id",
			req: &GetFilesInFolderRequest{
				FolderID: nil,
			},
			expectedFiles: []*File{
				{
					ID:        "file_id_1",
					Name:      "file_name_1",
					FolderID:  nil,
					CreatedAt: "file_created_at",
					UpdatedAt: "file_updated_at",
				},
				{
					ID:        "file_id_2",
					Name:      "file_name_2",
					FolderID:  nil,
					CreatedAt: "file_created_at",
					UpdatedAt: "file_updated_at",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES).Order("name", query.ASC)
				preparedQuery.Where().IS("folder_id", nil)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
							AddRow("file_id_1", "file_name_1", nil, "file_created_at", "file_updated_at").
							AddRow("file_id_2", "file_name_2", nil, "file_created_at", "file_updated_at"),
					)
			},
		},
		{
			name: "success with folder_id",
			req: &GetFilesInFolderRequest{
				FolderID: utils.MakePointer("folder_id"),
			},
			expectedFiles: []*File{
				{
					ID:        "file_id_1",
					Name:      "file_name_1",
					FolderID:  utils.MakePointer("folder_id"),
					CreatedAt: "file_created_at",
					UpdatedAt: "file_updated_at",
				},
				{
					ID:        "file_id_2",
					Name:      "file_name_2",
					FolderID:  utils.MakePointer("folder_id"),
					CreatedAt: "file_created_at",
					UpdatedAt: "file_updated_at",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES).Order("name", query.ASC)
				preparedQuery.Where().EQ("folder_id", "folder_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
							AddRow("file_id_1", "file_name_1", "folder_id", "file_created_at", "file_updated_at").
							AddRow("file_id_2", "file_name_2", "folder_id", "file_created_at", "file_updated_at"),
					)
			},
		},
		{
			name: "success with folder_id and order",
			req: &GetFilesInFolderRequest{
				FolderID: utils.MakePointer("folder_id"),
				Order: &Order{
					Column: utils.MakePointer("created_at"),
					Type:   utils.MakePointer(query.DESC),
				},
			},
			expectedFiles: []*File{
				{
					ID:        "file_id_1",
					Name:      "file_name_1",
					FolderID:  utils.MakePointer("folder_id"),
					CreatedAt: "file_created_at",
					UpdatedAt: "file_updated_at",
				},
				{
					ID:        "file_id_2",
					Name:      "file_name_2",
					FolderID:  utils.MakePointer("folder_id"),
					CreatedAt: "file_created_at",
					UpdatedAt: "file_updated_at",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES).Order("created_at", query.DESC)
				preparedQuery.Where().EQ("folder_id", "folder_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
							AddRow("file_id_1", "file_name_1", "folder_id", "file_created_at", "file_updated_at").
							AddRow("file_id_2", "file_name_2", "folder_id", "file_created_at", "file_updated_at"),
					)
			},
		},
		{
			name:          "empty request",
			req:           nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &GetFilesInFolderRequest{
				FolderID: utils.MakePointer("folder_id"),
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES).Order("name", query.ASC)
				preparedQuery.Where().EQ("folder_id", "folder_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			files, err := client.GetFilesInFolder(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, files)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFiles, files)
			}
		})
	}
}

func TestClient_CreateFile(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name          string
		req           *CreateRequest
		mockSetup     func()
		expectedFile  *File
		expectedError tiny_errors.ErrorHandler
	}{
		{
			name: "success with folder_id",
			req: &CreateRequest{
				Name:     "file_name",
				FolderID: utils.MakePointer("folder_id"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("name", "file_name")
				preparedQuery.Where().EQ("folder_id", "folder_id")

				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}),
					)

				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_FILE)).WithArgs(utils.MakePointer("folder_id"), "file_name").WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
						AddRow("file_id", "file_name", "folder_id", "file_created_at", "file_updated_at"),
				)
			},
			expectedFile: &File{
				ID:        "file_id",
				Name:      "file_name",
				FolderID:  utils.MakePointer("folder_id"),
				CreatedAt: "file_created_at",
				UpdatedAt: "file_updated_at",
			},
		},
		{
			name: "success without folder_id",
			req: &CreateRequest{
				Name:     "file_name",
				FolderID: nil,
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("name", "file_name")
				preparedQuery.Where().IS("folder_id", nil)

				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}),
					)

				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_FILE)).WithArgs(nil, "file_name").WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
						AddRow("file_id", "file_name", "folder_id", "file_created_at", "file_updated_at"),
				)
			},
			expectedFile: &File{
				ID:        "file_id",
				Name:      "file_name",
				FolderID:  utils.MakePointer("folder_id"),
				CreatedAt: "file_created_at",
				UpdatedAt: "file_updated_at",
			},
		},
		{
			name:          "empty request",
			req:           nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "empty name",
			req: &CreateRequest{
				Name:     "",
				FolderID: utils.MakePointer("folder_id"),
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("name", "required")),
		},
		{
			name: "file exists",
			req: &CreateRequest{
				Name:     "file_name",
				FolderID: utils.MakePointer("folder_id"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("name", "file_name")
				preparedQuery.Where().EQ("folder_id", "folder_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
							AddRow("file_id", "file_name", "folder_id", "file_created_at", "file_updated_at"),
					)
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file already exists")),
		},
		{
			name: "file existence error",
			req: &CreateRequest{
				Name:     "file_name",
				FolderID: utils.MakePointer("folder_id"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("name", "file_name")
				preparedQuery.Where().EQ("folder_id", "folder_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "sql create file error",
			req: &CreateRequest{
				Name:     "file_name",
				FolderID: utils.MakePointer("folder_id"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("name", "file_name")
				preparedQuery.Where().EQ("folder_id", "folder_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}),
					)

				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CREATE_FILE)).WithArgs(utils.MakePointer("folder_id"), "file_name").WillReturnError(assert.AnError)
			},
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			file, err := client.Create(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				if details := tt.expectedError.GetDetails(); details != nil {
					assert.Equal(t, details, err.GetDetails())
				}
				assert.Nil(t, file)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedFile, file)
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
			req: &DeleteRequest{
				ID: "123",
			},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FILE)).WithArgs("123").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "not found element, rows affected 0",
			req: &DeleteRequest{
				ID: "123",
			},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FILE)).WithArgs("123").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &DeleteRequest{
				ID: "123",
			},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FILE)).WithArgs("123").WillReturnError(assert.AnError)
			},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			removed, err := client.Delete(context.Background(), tt.req)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, removed)
		})
	}
}

func TestClient_EditFile(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *EditRequest
		mockSetup      func()
		expectedFolder *File
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &EditRequest{
				FileID: "123",
				Name:   "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FILE_EXISTS_BY_FOLDER_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(false),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_UPDATE_FILE)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).
						AddRow("123", "new_name", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z"),
				)
			},
			expectedFolder: &File{
				ID:        "123",
				Name:      "new_name",
				FolderID:  nil,
				CreatedAt: "2020-01-01T00:00:00Z",
				UpdatedAt: "2020-02-01T00:00:00Z",
			},
			expectedError: nil,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &EditRequest{
				FileID: "123",
				Name:   "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FILE_EXISTS_BY_FOLDER_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(false),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_UPDATE_FILE)).WithArgs("new_name", "123").WillReturnError(
					assert.AnError,
				)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "not found",
			req: &EditRequest{
				FileID: "123",
				Name:   "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FILE_EXISTS_BY_FOLDER_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(false),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_UPDATE_FILE)).WithArgs("new_name", "123").WillReturnError(
					sql.ErrNoRows,
				)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("not found")),
		},
		{
			name: "found the same name in the same folder",
			req: &EditRequest{
				FileID: "123",
				Name:   "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FILE_EXISTS_BY_FOLDER_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(true),
				)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file with such name already exists")),
		},
		{
			name: "empty id or name",
			req: &EditRequest{
				FileID: "",
				Name:   "",
			},
			mockSetup:      func() {},
			expectedFolder: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("file_id", "required"),
				tiny_errors.Detail("name", "required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			folder, err := client.Edit(context.Background(), tt.req)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				if tt.expectedError.GetDetails() != nil {
					assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedFolder, folder)
		})
	}
}

func TestClient_GetFile(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name          string
		req           *GetRequest
		mockSetup     func()
		expectedFile  *File
		expectedError tiny_errors.ErrorHandler
	}{
		{
			name: "success with root id",
			req: &GetRequest{
				ID: "root_id",
			},
			mockSetup: func() {
				expectedFile := &File{
					ID:        "root_id",
					Name:      "root_name",
					FolderID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				}
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("id", "root_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "folder_id", "created_at", "updated_at"}).AddRow(
						expectedFile.ID,
						expectedFile.Name,
						expectedFile.FolderID,
						expectedFile.CreatedAt,
						expectedFile.UpdatedAt,
					),
				)
			},
			expectedFile: &File{
				ID:        "root_id",
				Name:      "root_name",
				FolderID:  nil,
				CreatedAt: "2020-01-01T00:00:00Z",
				UpdatedAt: "2020-02-01T00:00:00Z",
			},
			expectedError: nil,
		},
		{
			name:          "empty request",
			req:           nil,
			mockSetup:     func() {},
			expectedFile:  nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &GetRequest{
				ID: "root_id",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FILES)
				preparedQuery.Where().EQ("id", "root_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnError(assert.AnError)
			},
			expectedFile:  nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			file, err := client.Get(context.Background(), tt.req)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedFile, file)
		})
	}
}
