package file_contents

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/query"
	"github.com/Moranilt/http-utils/tiny_errors"
)

type client struct {
	db *database.Client
}

type Client interface {
	// Create creates a new file content entry in the database.
	Create(ctx context.Context, req *CreateRequest) (*FileContent, tiny_errors.ErrorHandler)

	// GetMany retrieves multiple file content entries from the database.
	GetMany(ctx context.Context, req *GetManyRequest) ([]*FileContent, tiny_errors.ErrorHandler)

	// Edit updates an existing file content entry in the database.
	Edit(ctx context.Context, req *EditRequest) (*FileContent, tiny_errors.ErrorHandler)

	// Delete removes a file content entry from the database.
	Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler)
}

// New creates a new instance of the Client interface, which provides methods for
// interacting with file contents in a database.
func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

// TODO: store content as base64 string
func (c *client) Create(ctx context.Context, req *CreateRequest) (*FileContent, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
		{Name: "content", Value: req.Content},
		{Name: "version", Value: req.Version},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	row := c.db.QueryRowxContext(ctx, QUERY_GET_FILES_CONTENT_ID_BY_VERSION, req.FileID, req.Version)
	if row.Err() != nil && row.Err() != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var id string
	err := row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if len(id) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file content already exists"))
	}

	row = c.db.QueryRowxContext(ctx, QUERY_CREATE_CONTENT, req.FileID, req.Version, req.Content)
	if row.Err() != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var fileContent FileContent
	err = row.StructScan(&fileContent)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &fileContent, nil
}

func (c *client) GetMany(ctx context.Context, req *GetManyRequest) ([]*FileContent, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_FILE_CONTENTS)
	preparedQuery.Where().EQ("file_id", req.FileID)
	if req.Version != nil {
		preparedQuery.Where().EQ("version", req.Version)
	}

	files := make([]*FileContent, 0)
	err := c.db.SelectContext(ctx, &files, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return files, nil
}

func (c *client) Edit(ctx context.Context, req *EditRequest) (*FileContent, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "content_id", Value: req.FileContentID},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	if req.Version == nil && req.Content == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("version or content", "required"))
	}

	contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID)
	contentQuery.Where().EQ("id", req.FileContentID)
	row := c.db.QueryRowxContext(ctx, contentQuery.String())
	if row.Err() != nil && row.Err() != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var id string
	err := row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if len(id) == 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("file content does not exist"))
	}

	var queryUpdate strings.Builder
	queryUpdate.WriteString("UPDATE file_contents SET ")
	var setters []string
	setters = append(setters, "updated_at = now()")

	if req.Version != nil {
		setters = append(setters, fmt.Sprintf("version = '%s'", *req.Version))
	}

	if req.Content != nil {
		setters = append(setters, fmt.Sprintf("content = '%s'", *req.Content))
	}
	queryUpdate.WriteString(strings.Join(setters, ", "))
	queryUpdate.WriteString(fmt.Sprintf(" WHERE id = '%s'", req.FileContentID))
	queryUpdate.WriteString(" RETURNING id, file_id, version, content, created_at, updated_at")

	row = c.db.QueryRowxContext(ctx, queryUpdate.String())
	if row.Err() != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var fileContent FileContent
	err = row.StructScan(&fileContent)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &fileContent, nil
}

func (c *client) Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.ID},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return false, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	result, err := c.db.ExecContext(ctx, QUERY_DELETE_FILE_CONTENT, req.ID)
	if err != nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if affected == 0 {
		return false, nil
	}

	return true, nil
}
