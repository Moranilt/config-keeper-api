package file_contents

import (
	"context"
	"database/sql"

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

func (c *client) Create(ctx context.Context, req *CreateRequest) (*FileContent, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
		{Name: "content", Value: req.Content},
		{Name: "version", Value: req.Version},
		{Name: "format_id", Value: req.FormatID},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	var id string
	err := c.db.GetContext(ctx, &id, QUERY_GET_FILES_CONTENT_ID_BY_VERSION, req.FileID, req.Version)
	if err != nil && err != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if len(id) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file content already exists"))
	}

	base64Content := utils.StringToBase64(req.Content)

	var fileContent FileContent
	err = c.db.QueryRowxContext(ctx, QUERY_CREATE_CONTENT, req.FileID, req.Version, base64Content, req.FormatID).StructScan(&fileContent)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &fileContent, nil
}

func (c *client) GetMany(ctx context.Context, req *GetManyRequest) ([]*FileContent, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_FILE_CONTENTS).Where().EQ("file_id", req.FileID).Query()
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

	contentQuery := query.New(QUERY_GET_FILE_CONTENTS_ID).Where().EQ("id", req.FileContentID).Query()

	var id string
	err := c.db.GetContext(ctx, &id, contentQuery.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if len(id) == 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("file content does not exist"))
	}

	queryUpdate := query.New("UPDATE file_contents").Set("updated_at", "now()").
		Where().EQ("id", req.FileContentID).Query().
		Returning("id", "file_id", "version", "content", "created_at", "updated_at")

	if req.Version != nil {
		queryUpdate.Set("version", *req.Version)
	}

	if req.Content != nil {
		base64Content := utils.StringToBase64(*req.Content)
		queryUpdate.Set("content", base64Content)
	}

	var fileContent FileContent
	err = c.db.QueryRowxContext(ctx, queryUpdate.String()).StructScan(&fileContent)
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
