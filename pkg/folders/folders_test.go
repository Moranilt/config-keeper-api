package folders

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

func TestClient_NewFolder(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *CreateRequest
		mockSetup      func()
		expectedFolder *Folder
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &CreateRequest{
				Name:     "folder_name",
				ParentID: utils.MakePointer("folder_parent_id"),
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_INSERT_FOLDER)).
					WithArgs("folder_name", utils.MakePointer("folder_parent_id")).
					WillReturnRows(
						sqlMock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at"}).
							AddRow("folder_id", "folder_name", utils.MakePointer("folder_parent_id"), "folder_created_at", "folder_updated_at"),
					)
			},
			expectedFolder: &Folder{
				ID:        "folder_id",
				Name:      "folder_name",
				ParentID:  utils.MakePointer("folder_parent_id"),
				CreatedAt: "folder_created_at",
				UpdatedAt: "folder_updated_at",
			},
			expectedError: nil,
		},
		{
			name:           "body required",
			req:            nil,
			mockSetup:      func() {},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "sql error",
			req: &CreateRequest{
				Name:     "folder_name",
				ParentID: utils.MakePointer("folder_parent_id"),
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_INSERT_FOLDER)).
					WithArgs("folder_name", utils.MakePointer("folder_parent_id")).
					WillReturnError(assert.AnError)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			folder, err := client.Create(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedFolder, folder)
		})
	}
}

func TestClient_ExistsFolder(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *ExistsRequest
		setupMock      func()
		expectedExists bool
		expectedErr    tiny_errors.ErrorHandler
	}{
		{
			name: "success with parent_id",
			req: &ExistsRequest{
				Name:     utils.MakePointer("folder_name"),
				ParentID: utils.MakePointer("folder_parent_id"),
			},
			setupMock: func() {
				preparedQuery := query.New(QUERY_DEFAULT_SELECT_FOLDERS_ID)
				preparedQuery.Where().EQ("parent_id", utils.MakePointer("folder_parent_id")).EQ("name", utils.MakePointer("folder_name"))
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}).
						AddRow("folder_id"),
				)
			},
			expectedExists: true,
			expectedErr:    nil,
		},
		{
			name: "success without parent_id",
			req: &ExistsRequest{
				Name:     utils.MakePointer("folder_name"),
				ParentID: nil,
			},
			setupMock: func() {
				preparedQuery := query.New(QUERY_DEFAULT_SELECT_FOLDERS_ID)
				preparedQuery.Where().IS("parent_id", nil).EQ("name", utils.MakePointer("folder_name"))
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}).
						AddRow("folder_id"),
				)
			},
			expectedExists: true,
			expectedErr:    nil,
		},
		{
			name:           "body required",
			req:            nil,
			setupMock:      func() {},
			expectedExists: false,
			expectedErr:    tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &ExistsRequest{
				Name:     utils.MakePointer("folder_name"),
				ParentID: utils.MakePointer("folder_parent_id"),
			},
			setupMock: func() {
				preparedQuery := query.New(QUERY_DEFAULT_SELECT_FOLDERS_ID)
				preparedQuery.Where().EQ("parent_id", utils.MakePointer("folder_parent_id")).EQ("name", utils.MakePointer("folder_name"))
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnError(
					assert.AnError,
				)
			},
			expectedExists: false,
			expectedErr:    tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "sql no rows",
			req: &ExistsRequest{
				Name:     utils.MakePointer("folder_name"),
				ParentID: utils.MakePointer("folder_parent_id"),
			},
			setupMock: func() {
				preparedQuery := query.New(QUERY_DEFAULT_SELECT_FOLDERS_ID)
				preparedQuery.Where().EQ("parent_id", utils.MakePointer("folder_parent_id")).EQ("name", utils.MakePointer("folder_name"))
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id"}),
				)
			},
			expectedExists: false,
			expectedErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			exists, err := client.Exists(context.Background(), tt.req)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedErr.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedExists, exists)
		})
	}
}

func TestClient_GetFolder(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *GetRequest
		expectedFolder *FolderWithPath
		expectedError  tiny_errors.ErrorHandler
		mockSetup      func()
	}{
		{
			name: "success with root id",
			req: &GetRequest{
				ID: "root_id",
			},
			expectedFolder: &FolderWithPath{
				Folder: Folder{
					ID:        "root_id",
					Name:      "root_name",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
				Path: "root_name",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FOLDER_WITH_PATH)
				preparedQuery.Where().EQ("id", "root_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "parent_id", "path", "created_at", "updated_at"}).AddRow(
						"root_id",
						"root_name",
						nil,
						"root_name",
						"2020-01-01T00:00:00Z",
						"2020-02-01T00:00:00Z",
					),
				)
			},
		},
		{
			name:           "empty request",
			req:            nil,
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &GetRequest{
				ID: "root_id",
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FOLDER_WITH_PATH)
				preparedQuery.Where().EQ("id", "root_id")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnError(
					assert.AnError,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			folder, err := client.Get(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedFolder, folder)
		})
	}
}

func TestClient_GetFolders(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *GetManyRequest
		expectedResult []*Folder
		expectedError  tiny_errors.ErrorHandler
		mockSetup      func()
	}{
		{
			name: "success with empty parent_id",
			req: &GetManyRequest{
				ParentID: nil,
			},
			expectedResult: []*Folder{
				{
					ID:        "folder_id_1",
					Name:      "folder_name_1",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
				{
					ID:        "folder_id_2",
					Name:      "folder_name_2",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FOLDERS).Order("name", query.ASC)
				preparedQuery.Where().IS("parent_id", nil)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at"}).
						AddRow("folder_id_1", "folder_name_1", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z").
						AddRow("folder_id_2", "folder_name_2", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z"),
				)
			},
		},
		{
			name: "success with parent_id",
			req: &GetManyRequest{
				ParentID: utils.MakePointer("123"),
			},
			expectedResult: []*Folder{
				{
					ID:        "folder_id_1",
					Name:      "folder_name_1",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
				{
					ID:        "folder_id_2",
					Name:      "folder_name_2",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FOLDERS).Order("name", query.ASC)
				preparedQuery.Where().EQ("parent_id", "123")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at"}).
						AddRow("folder_id_1", "folder_name_1", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z").
						AddRow("folder_id_2", "folder_name_2", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z"),
				)
			},
		},
		{
			name: "success with parent_id and order",
			req: &GetManyRequest{
				ParentID: utils.MakePointer("123"),
				Order: &Order{
					Column: utils.MakePointer("created_at"),
					Type:   utils.MakePointer(query.DESC),
				},
			},
			expectedResult: []*Folder{
				{
					ID:        "folder_id_1",
					Name:      "folder_name_1",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
				{
					ID:        "folder_id_2",
					Name:      "folder_name_2",
					ParentID:  nil,
					CreatedAt: "2020-01-01T00:00:00Z",
					UpdatedAt: "2020-02-01T00:00:00Z",
				},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FOLDERS).Order("created_at", query.DESC)
				preparedQuery.Where().EQ("parent_id", "123")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at"}).
						AddRow("folder_id_1", "folder_name_1", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z").
						AddRow("folder_id_2", "folder_name_2", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z"),
				)
			},
		},
		{
			name:           "empty request",
			req:            nil,
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required")),
		},
		{
			name: "sql error",
			req: &GetManyRequest{
				ParentID: utils.MakePointer("123"),
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_FOLDERS)
				preparedQuery.Where().EQ("parent_id", "123")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WillReturnError(assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			folders, err := client.GetMany(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Nil(t, folders)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, folders)
				assert.Equal(t, len(tt.expectedResult), len(folders))
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
			req:  &DeleteRequest{ID: "123"},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FOLDER)).WithArgs("123").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "not found element, rows affected 0",
			req:  &DeleteRequest{ID: "123"},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FOLDER)).WithArgs("123").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name:           "empty request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired, tiny_errors.Message("body required"))},
		{
			name: "sql error",
			req:  &DeleteRequest{ID: "123"},
			mockSetup: func() {
				sqlMock.ExpectExec(regexp.QuoteMeta(QUERY_DELETE_FOLDER)).WithArgs("123").WillReturnError(assert.AnError)
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

func TestClient_EditFolder(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *EditRequest
		mockSetup      func()
		expectedFolder *Folder
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &EditRequest{
				ID:   "123",
				Name: "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FOLDER_EXISTS_BY_PARENT_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(false),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_UPDATE_FOLDER)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"id", "name", "parent_id", "created_at", "updated_at"}).
						AddRow("123", "new_name", nil, "2020-01-01T00:00:00Z", "2020-02-01T00:00:00Z"),
				)
			},
			expectedFolder: &Folder{
				ID:        "123",
				Name:      "new_name",
				ParentID:  nil,
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
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "sql error",
			req: &EditRequest{
				ID:   "123",
				Name: "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FOLDER_EXISTS_BY_PARENT_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(false),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_UPDATE_FOLDER)).WithArgs("new_name", "123").WillReturnError(assert.AnError)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "not found",
			req: &EditRequest{
				ID:   "123",
				Name: "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FOLDER_EXISTS_BY_PARENT_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(false),
				)
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_UPDATE_FOLDER)).WithArgs("new_name", "123").WillReturnError(sql.ErrNoRows)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_NotFound),
		},
		{
			name: "found the same name in the same folder",
			req: &EditRequest{
				ID:   "123",
				Name: "new_name",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_FOLDER_EXISTS_BY_PARENT_ID_AND_NAME)).WithArgs("new_name", "123").WillReturnRows(
					sqlMock.NewRows([]string{"exists"}).AddRow(true),
				)
			},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("folder with such name already exists")),
		},
		{
			name: "empty id or name",
			req: &EditRequest{
				ID:   "",
				Name: "",
			},
			mockSetup:      func() {},
			expectedFolder: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_NotValid),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			folder, err := client.Edit(context.Background(), tt.req)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedFolder, folder)
		})
	}
}
