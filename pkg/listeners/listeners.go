package listeners

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
	// Create creates a new listener in the database
	Create(ctx context.Context, req *CreateRequest) (*Listener, tiny_errors.ErrorHandler)

	// GetMany retrieves multiple listeners from the database
	GetMany(ctx context.Context, req *GetManyRequest) ([]*Listener, tiny_errors.ErrorHandler)

	// Get retrieves a single listener from the database
	Get(ctx context.Context, req *GetRequest) (*Listener, tiny_errors.ErrorHandler)

	// Delete removes a listener from the database
	Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler)

	// Edit updates an existing listener in the database
	Edit(ctx context.Context, req *EditRequest) (*Listener, tiny_errors.ErrorHandler)
}

// New creates a new Client implementation backed by the provided database.Client.
func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

func (c *client) Create(ctx context.Context, req *CreateRequest) (*Listener, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
		{Name: "callback_endpoint", Value: req.CallbackEndpoint},
		{Name: "name", Value: req.Name},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	row := c.db.QueryRowxContext(ctx, QUERY_CREATE_LISTENER, req.FileID, req.CallbackEndpoint, req.Name)
	if row.Err() != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var listener Listener
	err := row.StructScan(&listener)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &listener, nil
}

func (c *client) GetMany(ctx context.Context, req *GetManyRequest) ([]*Listener, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_LISTENERS).Where().EQ("file_id", req.FileID).Query().Order("name", "asc")

	listeners := make([]*Listener, 0)
	err := c.db.SelectContext(ctx, &listeners, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return listeners, nil
}

func (c *client) Get(ctx context.Context, req *GetRequest) (*Listener, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_LISTENERS).Where().EQ("id", req.ID).Query()
	var listener Listener
	err := c.db.GetContext(ctx, &listener, preparedQuery.String())
	if err != nil && err != sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if err == sql.ErrNoRows {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound)
	}

	return &listener, nil
}

func (c *client) Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	result, err := c.db.ExecContext(ctx, QUERY_DELETE_LISTENER, req.ID)
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

func (c *client) Edit(ctx context.Context, req *EditRequest) (*Listener, tiny_errors.ErrorHandler) {
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

	if req.CallbackEndpoint == nil && req.Name == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("name or callback_endpoint", "required"))
	}

	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}
	defer tx.Rollback()

	var listener Listener
	preparedQuery := query.New(QUERY_GET_LISTENERS).Where().EQ("id", req.ID).Query()
	err = tx.QueryRowxContext(ctx, preparedQuery.String()).StructScan(&listener)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.Message("listener does not exist"))
		}
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	updateQuery := buildUpdateQuery(req)
	err = tx.QueryRowxContext(ctx, updateQuery).StructScan(&listener)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	if err := tx.Commit(); err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &listener, nil
}

func buildUpdateQuery(req *EditRequest) string {
	queryBuilder := query.New("UPDATE listeners").Set("updated_at", "now()").Where().EQ("id", req.ID).Query().
		Returning("id", "file_id", "callback_endpoint", "name", "created_at", "updated_at")

	if req.Name != nil {
		queryBuilder.Set("name", *req.Name)
	}
	if req.CallbackEndpoint != nil {
		queryBuilder.Set("callback_endpoint", *req.CallbackEndpoint)
	}

	return queryBuilder.String()
}
