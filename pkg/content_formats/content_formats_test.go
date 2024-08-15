package content_formats

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/http-utils/clients/database"
	database_mock "github.com/Moranilt/http-utils/clients/database/mock"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/stretchr/testify/assert"
)

func TestGetMany(t *testing.T) {
	tiny_errors.Init(custom_errors.ERRORS)
	mockDb, sqlMock := database_mock.NewSQlMock(t)
	client := New(&database.Client{mockDb})

	tests := []struct {
		name           string
		mockDBResponse []*ContentFormat
		mockDBError    error
		expectedResult []*ContentFormat
		expectedError  tiny_errors.ErrorHandler
	}{
		{
			name: "Successful retrieval",
			mockDBResponse: []*ContentFormat{
				{ID: "1", Name: "Format1"},
				{ID: "2", Name: "Format2"},
			},
			mockDBError:    nil,
			expectedResult: []*ContentFormat{{ID: "1", Name: "Format1"}, {ID: "2", Name: "Format2"}},
			expectedError:  nil,
		},
		{
			name:           "Database error",
			mockDBResponse: nil,
			mockDBError:    errors.New("database error"),
			expectedResult: nil,
			expectedError:  tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message("database error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the expected query
			rows := sqlmock.NewRows([]string{"id", "name"})
			for _, cf := range tt.mockDBResponse {
				rows.AddRow(cf.ID, cf.Name)
			}

			if tt.mockDBError != nil {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GET_FORMATS)).WillReturnError(tt.mockDBError)
			} else {
				sqlMock.ExpectQuery(regexp.QuoteMeta(QUERY_GET_FORMATS)).WillReturnRows(rows)
			}

			// Call the function
			result, err := client.GetMany(context.Background())

			// Check the results
			assert.Equal(t, tt.expectedResult, result)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Ensure all expectations were met
			if err := sqlMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
