package files

import (
	"context"
	"database/sql"
	"net/http"

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
	// GetMany retrieves multiple files based on the provided request parameters.
	GetMany(ctx context.Context, req *GetManyRequest) ([]*File, tiny_errors.ErrorHandler)

	// Create adds a new file to the system.
	Create(ctx context.Context, req *CreateRequest) (*File, tiny_errors.ErrorHandler)

	// Delete removes a file from the system.
	Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler)

	// Edit modifies an existing file.
	Edit(ctx context.Context, req *EditRequest) (*File, tiny_errors.ErrorHandler)

	// Get retrieves a single file.
	Get(ctx context.Context, req *GetRequest) (*File, tiny_errors.ErrorHandler)
}

// New creates a new instance of the Client interface, which provides methods for
// interacting with files in the system. The Client implementation is backed by
// the provided database.Client.
func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

func (c *client) GetMany(ctx context.Context, req *GetManyRequest) ([]*File, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_FILES)
	if req.Order != nil && req.Order.Column != nil && req.Order.Type != nil {
		order := query.ASC
		if *req.Order.Type == query.DESC || *req.Order.Type == query.ASC {
			order = *req.Order.Type
		}
		preparedQuery = preparedQuery.Order(*req.Order.Column, order)
	} else {
		preparedQuery = preparedQuery.Order("name", query.ASC)
	}

	if req.FolderID == nil {
		preparedQuery.Where().IS("folder_id", nil)
	} else {
		preparedQuery.Where().EQ("folder_id", req.FolderID)
	}

	files := make([]*File, 0)
	err := c.db.SelectContext(ctx, &files, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return files, nil
}

func (c *client) Create(ctx context.Context, req *CreateRequest) (*File, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "name", Value: req.Name},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New(QUERY_GET_FILES)
	preparedQuery.Where().EQ("name", req.Name)
	if req.FolderID != nil {
		preparedQuery.Where().EQ("folder_id", req.FolderID)
	} else {
		preparedQuery.Where().IS("folder_id", nil)
	}

	row := c.db.QueryRowxContext(ctx, preparedQuery.String())
	if row.Err() != nil && row.Err() != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var existsFile File
	err := row.StructScan(&existsFile)
	if err != nil && err != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if existsFile.ID != "" {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file already exists"))
	}

	row = c.db.QueryRowxContext(ctx, QUERY_CREATE_FILE, req.FolderID, req.Name)
	if row.Err() != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var file File
	err = row.StructScan(&file)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &file, nil
}

func (c *client) Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "id", Value: req.ID},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return false, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	result, err := c.db.ExecContext(ctx, QUERY_DELETE_FILE, req.ID)
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

func (c *client) Edit(ctx context.Context, req *EditRequest) (*File, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{
			Name:  "file_id",
			Value: req.FileID,
		},
		{
			Name:  "name",
			Value: req.Name,
		},
	}

	errFields := utils.ValidateRequiredFields(requiredFields)
	if errFields != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, errFields...)
	}

	row := c.db.QueryRowxContext(
		ctx,
		QUERY_CHECK_FILE_EXISTS_BY_FOLDER_ID_AND_NAME,
		req.Name,
		req.FileID,
	)
	if row.Err() != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Marshal, tiny_errors.Message(err.Error()))
	}

	if exists {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("file with such name already exists"))
	}

	row = c.db.QueryRowxContext(ctx, QUERY_UPDATE_FILE, req.Name, req.FileID)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.HTTPStatus(http.StatusNotFound))
		}
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var folder File
	err = row.StructScan(&folder)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Marshal, tiny_errors.Message(err.Error()))
	}

	return &folder, nil
}

func (c *client) Get(ctx context.Context, req *GetRequest) (*File, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "id", Value: req.ID},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}
	preparedQuery := query.New(QUERY_GET_FILES)
	preparedQuery.Where().EQ("id", req.ID)
	var file File
	err := c.db.GetContext(ctx, &file, preparedQuery.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.HTTPStatus(http.StatusNotFound))
		}
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &file, nil
}
