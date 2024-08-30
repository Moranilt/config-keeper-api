package aliases

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

func TestClient_CreateFileContent(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *CreateRequest
		mockSetup      func()
		expectedResult *Alias
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success",
			req: &CreateRequest{
				Key:   "key1",
				Value: "value1",
				Color: "color1",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_CREATE_ALIAS).
					InsertColumns("key", "value", "color").
					Values("?", "?", "?").
					Returning("id", "key", "value", "color", "created_at")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WithArgs("key1", "value1", "color1").WillReturnRows(
					sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}).
						AddRow("alias_id", "key1", "value1", "color1", "alias_created_at"),
				)
			},
			expectedResult: &Alias{
				ID:        "alias_id",
				Key:       "key1",
				Value:     "value1",
				Color:     "color1",
				CreatedAt: "alias_created_at",
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
				Key:   "key1",
				Value: "value1",
				Color: "color1",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_CREATE_ALIAS).
					InsertColumns("key", "value", "color").
					Values("?", "?", "?").
					Returning("id", "key", "value", "color", "created_at")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).WithArgs("key1", "value1", "color1").WillReturnError(
					assert.AnError,
				)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "empty fields in request",
			req: &CreateRequest{
				Key:   "",
				Value: "",
				Color: "",
			},
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("key", "required"),
				tiny_errors.Detail("value", "required"),
				tiny_errors.Detail("color", "required"),
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
func TestClient_Exists(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *ExistsRequest
		mockSetup      func()
		expectedResult bool
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "alias exists",
			req: &ExistsRequest{
				Key:   "existing_key",
				Value: "existing_value",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_ALIAS_EXISTS)).
					WithArgs("existing_key", "existing_value").
					WillReturnRows(sqlMock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "alias does not exist",
			req: &ExistsRequest{
				Key:   "non_existing_key",
				Value: "non_existing_value",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_ALIAS_EXISTS)).
					WithArgs("non_existing_key", "non_existing_value").
					WillReturnRows(sqlMock.NewRows([]string{"exists"}).AddRow(false))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "empty key",
			req: &ExistsRequest{
				Key:   "",
				Value: "some_value",
			},
			mockSetup:      func() {},
			expectedResult: false,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("key", "required"),
			),
		},
		{
			name: "empty value",
			req: &ExistsRequest{
				Key:   "some_key",
				Value: "",
			},
			mockSetup:      func() {},
			expectedResult: false,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("value", "required"),
			),
		},
		{
			name:           "nil request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "database error",
			req: &ExistsRequest{
				Key:   "error_key",
				Value: "error_value",
			},
			mockSetup: func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_CHECK_ALIAS_EXISTS)).
					WithArgs("error_key", "error_value").
					WillReturnError(assert.AnError)
			},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			exists, err := client.Exists(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResult, exists)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestClient_ExistsInFile(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *ExistsInFileRequest
		mockSetup      func()
		expectedResult []string
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success with multiple aliases",
			req: &ExistsInFileRequest{
				FileID:  "file1",
				Aliases: []string{"alias1", "alias2", "alias3"},
			},
			mockSetup: func() {
				preparedQuery := query.New("SELECT aliases.id FROM aliases").
					InnerJoin("files_aliases as fa", "fa.alias_id=aliases.id").
					InnerJoin("files", "fa.file_id=files.id").
					Where().IN("aliases.id", []string{"alias1", "alias2", "alias3"}).EQ("files.id", "file1").Query()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id"}).AddRow("alias1").AddRow("alias3"))
			},
			expectedResult: []string{"alias1", "alias3"},
			expectedError:  nil,
		},
		{
			name: "no aliases found",
			req: &ExistsInFileRequest{
				FileID:  "file2",
				Aliases: []string{"alias4", "alias5"},
			},
			mockSetup: func() {
				preparedQuery := query.New("SELECT aliases.id FROM aliases").
					InnerJoin("files_aliases as fa", "fa.alias_id=aliases.id").
					InnerJoin("files", "fa.file_id=files.id").
					Where().IN("aliases.id", []string{"alias4", "alias5"}).EQ("files.id", "file2").Query()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id"}))
			},
			expectedResult: []string{},
			expectedError:  nil,
		},
		{
			name: "empty aliases list",
			req: &ExistsInFileRequest{
				FileID:  "file3",
				Aliases: []string{},
			},
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:           "nil request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "missing file_id",
			req: &ExistsInFileRequest{
				FileID:  "",
				Aliases: []string{"alias1"},
			},
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("file_id", "required"),
			),
		},
		{
			name: "database error",
			req: &ExistsInFileRequest{
				FileID:  "file4",
				Aliases: []string{"alias6"},
			},
			mockSetup: func() {
				preparedQuery := query.New("SELECT aliases.id FROM aliases").
					InnerJoin("files_aliases as fa", "fa.alias_id=aliases.id").
					InnerJoin("files", "fa.file_id=files.id").
					Where().IN("aliases.id", []string{"alias6"}).EQ("files.id", "file4").Query()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.ExistsInFile(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
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
			name: "successful deletion",
			req: &DeleteRequest{
				AliasID: "alias1",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_DELETE_ALIAS).Where().EQ("id", "?").Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WithArgs("alias1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "alias not found",
			req: &DeleteRequest{
				AliasID: "non_existing_alias",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_DELETE_ALIAS).Where().EQ("id", "?").Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WithArgs("non_existing_alias").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "database error on exec",
			req: &DeleteRequest{
				AliasID: "error_alias",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_DELETE_ALIAS).Where().EQ("id", "?").Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WithArgs("error_alias").
					WillReturnError(assert.AnError)
			},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "error on rows affected",
			req: &DeleteRequest{
				AliasID: "rows_error_alias",
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_DELETE_ALIAS).Where().EQ("id", "?").Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WithArgs("rows_error_alias").
					WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name:           "nil request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: false,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "empty alias_id",
			req: &DeleteRequest{
				AliasID: "",
			},
			mockSetup:      func() {},
			expectedResult: false,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("alias_id", "required"),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.Delete(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResult, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestClient_GetMany(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *GetManyRequest
		mockSetup      func()
		expectedResult []*Alias
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success with default limit and offset",
			req:  &GetManyRequest{},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit(DEFAULT_LIMIT).Offset(DEFAULT_OFFSET).Order("key", query.ASC)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}).
						AddRow("1", "key1", "value1", "color1", "2023-01-01").
						AddRow("2", "key2", "value2", "color2", "2023-01-02"))
			},
			expectedResult: []*Alias{
				{ID: "1", Key: "key1", Value: "value1", Color: "color1", CreatedAt: "2023-01-01"},
				{ID: "2", Key: "key2", Value: "value2", Color: "color2", CreatedAt: "2023-01-02"},
			},
			expectedError: nil,
		},
		{
			name: "success with custom limit, offset, and order",
			req: &GetManyRequest{
				Limit:     utils.MakePointer("5"),
				Offset:    utils.MakePointer("10"),
				OrderBy:   utils.MakePointer("value"),
				OrderType: utils.MakePointer("DESC"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit("5").Offset("10").Order("value", query.DESC)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}).
						AddRow("3", "key3", "value3", "color3", "2023-01-03"))
			},
			expectedResult: []*Alias{
				{ID: "3", Key: "key3", Value: "value3", Color: "color3", CreatedAt: "2023-01-03"},
			},
			expectedError: nil,
		},
		{
			name: "success with key filter",
			req: &GetManyRequest{
				Key: utils.MakePointer("specific_key"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit(DEFAULT_LIMIT).Offset(DEFAULT_OFFSET).Order("key", query.ASC).Where().EQ("key", "specific_key").Query()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}).
						AddRow("4", "specific_key", "value4", "color4", "2023-01-04"))
			},
			expectedResult: []*Alias{
				{ID: "4", Key: "specific_key", Value: "value4", Color: "color4", CreatedAt: "2023-01-04"},
			},
			expectedError: nil,
		},
		{
			name: "success with value filter",
			req: &GetManyRequest{
				Value: utils.MakePointer("specific_value"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit(DEFAULT_LIMIT).Offset(DEFAULT_OFFSET).Order("key", query.ASC).Where().EQ("value", "specific_value").Query()
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}).
						AddRow("5", "key5", "specific_value", "color5", "2023-01-05"))
			},
			expectedResult: []*Alias{
				{ID: "5", Key: "key5", Value: "specific_value", Color: "color5", CreatedAt: "2023-01-05"},
			},
			expectedError: nil,
		},
		{
			name: "no results",
			req:  &GetManyRequest{},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit(DEFAULT_LIMIT).Offset(DEFAULT_OFFSET).Order("key", query.ASC)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}))
			},
			expectedResult: []*Alias{},
			expectedError:  nil,
		},
		{
			name: "database error",
			req:  &GetManyRequest{},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit(DEFAULT_LIMIT).Offset(DEFAULT_OFFSET).Order("key", query.ASC)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name:           "nil request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "invalid order type",
			req: &GetManyRequest{
				OrderBy:   utils.MakePointer("key"),
				OrderType: utils.MakePointer("INVALID"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_GET_ALIASES).Limit(DEFAULT_LIMIT).Offset(DEFAULT_OFFSET).Order("key", query.ASC)
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at"}).
						AddRow("6", "key6", "value6", "color6", "2023-01-06"))
			},
			expectedResult: []*Alias{
				{ID: "6", Key: "key6", Value: "value6", Color: "color6", CreatedAt: "2023-01-06"},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.GetMany(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
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
		mockSetup      func()
		expectedResult *Alias
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success update all fields",
			req: &EditRequest{
				AliasID: "alias1",
				Key:     utils.MakePointer("new_key"),
				Value:   utils.MakePointer("new_value"),
				Color:   utils.MakePointer("new_color"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_UPDATE_ALIAS).
					Set("updated_at", "now()").
					Set("color", "new_color").
					Set("key", "new_key").
					Set("value", "new_value").
					Where().EQ("id", "alias1").Query().
					Returning("id", "key", "value", "color", "created_at", "updated_at")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at", "updated_at"}).
						AddRow("alias1", "new_key", "new_value", "new_color", "2023-01-01", "2023-01-02"))
			},
			expectedResult: &Alias{
				ID:        "alias1",
				Key:       "new_key",
				Value:     "new_value",
				Color:     "new_color",
				CreatedAt: "2023-01-01",
				UpdatedAt: "2023-01-02",
			},
			expectedError: nil,
		},
		{
			name: "success update single field",
			req: &EditRequest{
				AliasID: "alias2",
				Key:     utils.MakePointer("updated_key"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_UPDATE_ALIAS).
					Set("updated_at", "now()").
					Set("key", "updated_key").
					Where().EQ("id", "alias2").Query().
					Returning("id", "key", "value", "color", "created_at", "updated_at")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnRows(sqlMock.NewRows([]string{"id", "key", "value", "color", "created_at", "updated_at"}).
						AddRow("alias2", "updated_key", "old_value", "old_color", "2023-01-01", "2023-01-03"))
			},
			expectedResult: &Alias{
				ID:        "alias2",
				Key:       "updated_key",
				Value:     "old_value",
				Color:     "old_color",
				CreatedAt: "2023-01-01",
				UpdatedAt: "2023-01-03",
			},
			expectedError: nil,
		},
		{
			name: "error no fields to update",
			req: &EditRequest{
				AliasID: "alias3",
			},
			mockSetup:      func() {},
			expectedResult: nil,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("key, value or color", "required"),
			),
		},
		{
			name: "error alias not found",
			req: &EditRequest{
				AliasID: "non_existent_alias",
				Key:     utils.MakePointer("new_key"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_UPDATE_ALIAS).
					Set("updated_at", "now()").
					Set("key", "new_key").
					Where().EQ("id", "non_existent_alias").Query().
					Returning("id", "key", "value", "color", "created_at", "updated_at")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(sql.ErrNoRows)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(sql.ErrNoRows.Error())),
		},
		{
			name: "error database error",
			req: &EditRequest{
				AliasID: "alias4",
				Value:   utils.MakePointer("new_value"),
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_UPDATE_ALIAS).
					Set("updated_at", "now()").
					Set("value", "new_value").
					Where().EQ("id", "alias4").Query().
					Returning("id", "key", "value", "color", "created_at", "updated_at")
				sqlMock.ExpectQuery(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.Edit(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
func TestClient_AddToFile(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *AddToFileRequest
		mockSetup      func()
		expectedResult int
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success add multiple aliases",
			req: &AddToFileRequest{
				FileID:  "file1",
				Aliases: []string{"alias1", "alias2", "alias3"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_ADD_TO_FILE).InsertColumns("file_id", "alias_id").
					Values("file1", "alias1").
					Values("file1", "alias2").
					Values("file1", "alias3")
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			expectedResult: 3,
			expectedError:  nil,
		},
		{
			name: "success add single alias",
			req: &AddToFileRequest{
				FileID:  "file2",
				Aliases: []string{"alias4"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_ADD_TO_FILE).InsertColumns("file_id", "alias_id").
					Values("file2", "alias4")
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedResult: 1,
			expectedError:  nil,
		},
		{
			name:           "nil request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: 0,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "missing file_id",
			req: &AddToFileRequest{
				FileID:  "",
				Aliases: []string{"alias5"},
			},
			mockSetup:      func() {},
			expectedResult: 0,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("file_id", "required"),
			),
		},
		{
			name: "empty aliases list",
			req: &AddToFileRequest{
				FileID:  "file3",
				Aliases: []string{},
			},
			mockSetup:      func() {},
			expectedResult: 0,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("aliases", "required"),
			),
		},
		{
			name: "database exec error",
			req: &AddToFileRequest{
				FileID:  "file4",
				Aliases: []string{"alias6"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_ADD_TO_FILE).InsertColumns("file_id", "alias_id").
					Values("file4", "alias6")
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
			expectedResult: 0,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "rows affected error",
			req: &AddToFileRequest{
				FileID:  "file5",
				Aliases: []string{"alias7"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_ADD_TO_FILE).InsertColumns("file_id", "alias_id").
					Values("file5", "alias7")
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			expectedResult: 0,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.AddToFile(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResult, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestClient_RemoveFromFile(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		req            *RemoveFromFileRequest
		mockSetup      func()
		expectedResult int
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "success remove multiple aliases",
			req: &RemoveFromFileRequest{
				FileID:  "file1",
				Aliases: []string{"alias1", "alias2", "alias3"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_REMOVE_FROM_FILE).Where().EQ("file_id", "file1").IN("alias_id", []string{"alias1", "alias2", "alias3"}).Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewResult(0, 3))
			},
			expectedResult: 3,
			expectedError:  nil,
		},
		{
			name: "success remove single alias",
			req: &RemoveFromFileRequest{
				FileID:  "file2",
				Aliases: []string{"alias4"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_REMOVE_FROM_FILE).Where().EQ("file_id", "file2").IN("alias_id", []string{"alias4"}).Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectedResult: 1,
			expectedError:  nil,
		},
		{
			name:           "nil request",
			req:            nil,
			mockSetup:      func() {},
			expectedResult: 0,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_BodyRequired),
		},
		{
			name: "missing file_id",
			req: &RemoveFromFileRequest{
				FileID:  "",
				Aliases: []string{"alias5"},
			},
			mockSetup:      func() {},
			expectedResult: 0,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("file_id", "required"),
			),
		},
		{
			name: "empty aliases list",
			req: &RemoveFromFileRequest{
				FileID:  "file3",
				Aliases: []string{},
			},
			mockSetup:      func() {},
			expectedResult: 0,
			expectedError: tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD,
				tiny_errors.Detail("aliases", "required"),
			),
		},
		{
			name: "database exec error",
			req: &RemoveFromFileRequest{
				FileID:  "file4",
				Aliases: []string{"alias6"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_REMOVE_FROM_FILE).Where().EQ("file_id", "file4").IN("alias_id", []string{"alias6"}).Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnError(assert.AnError)
			},
			expectedResult: 0,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "rows affected error",
			req: &RemoveFromFileRequest{
				FileID:  "file5",
				Aliases: []string{"alias7"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_REMOVE_FROM_FILE).Where().EQ("file_id", "file5").IN("alias_id", []string{"alias7"}).Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewErrorResult(assert.AnError))
			},
			expectedResult: 0,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(assert.AnError.Error())),
		},
		{
			name: "no aliases removed",
			req: &RemoveFromFileRequest{
				FileID:  "file6",
				Aliases: []string{"non_existent_alias"},
			},
			mockSetup: func() {
				preparedQuery := query.New(QUERY_REMOVE_FROM_FILE).Where().EQ("file_id", "file6").IN("alias_id", []string{"non_existent_alias"}).Query()
				sqlMock.ExpectExec(regexp.QuoteMeta(preparedQuery.String())).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedResult: 0,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := client.RemoveFromFile(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResult, result)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.GetCode(), err.GetCode())
				assert.Equal(t, tt.expectedError.GetMessage(), err.GetMessage())
				assert.Equal(t, tt.expectedError.GetDetails(), err.GetDetails())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
