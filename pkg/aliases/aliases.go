package aliases

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/query"
	"github.com/Moranilt/http-utils/tiny_errors"
)

const (
	DEFAULT_LIMIT  = "10"
	DEFAULT_OFFSET = "0"
)

type client struct {
	db *database.Client
}

type Client interface {
	// Creates a new alias in the database
	Create(ctx context.Context, req *CreateRequest) (*Alias, tiny_errors.ErrorHandler)

	// Checks if an alias exists in the database
	Exists(ctx context.Context, req *ExistsRequest) (bool, tiny_errors.ErrorHandler)

	// Checks if aliases exist in a specific file
	ExistsInFile(ctx context.Context, req *ExistsInFileRequest) ([]string, tiny_errors.ErrorHandler)

	// Removes an alias from the database
	Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler)

	// Retrieves multiple aliases from the database
	GetMany(ctx context.Context, req *GetManyRequest) ([]*Alias, tiny_errors.ErrorHandler)

	// Retrieves a single alias from the database based on the provided request
	Get(ctx context.Context, req *GetRequest) (*Alias, tiny_errors.ErrorHandler)

	// Updates an existing alias in the database
	Edit(ctx context.Context, req *EditRequest) (*Alias, tiny_errors.ErrorHandler)

	// Adds aliases to a specific file
	AddToFile(ctx context.Context, req *AddToFileRequest) (int, tiny_errors.ErrorHandler)

	// Removes aliases from a specific file
	RemoveFromFile(ctx context.Context, req *RemoveFromFileRequest) (int, tiny_errors.ErrorHandler)

	// Retrieves aliases from a specific file
	GetFileAliases(ctx context.Context, req *GetFileAliasesRequest) ([]*Alias, tiny_errors.ErrorHandler)

	// Retrieves aliases of all provided file ids
	GetFilesAliasesManyToMany(ctx context.Context, req *GetFilesAliasesManyToManyRequest) ([]*AliasWithFileId, tiny_errors.ErrorHandler)
}

// New creates a new instance of the Client interface, which provides methods for
// interacting with alias contents in a database.
func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

func (c *client) Create(ctx context.Context, req *CreateRequest) (*Alias, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "key", Value: req.Key},
		{Name: "value", Value: req.Value},
		{Name: "color", Value: req.Color},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New(QUERY_CREATE_ALIAS).
		InsertColumns("key", "value", "color").
		Values(req.Key, req.Value, req.Color).
		Returning("id", "key", "value", "color", "created_at", "updated_at")

	var alias Alias
	err := c.db.QueryRowxContext(ctx, preparedQuery.String()).StructScan(&alias)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &alias, nil
}

func (c *client) Exists(ctx context.Context, req *ExistsRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "key", Value: req.Key},
		{Name: "value", Value: req.Value},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return false, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	var exists bool
	err := c.db.QueryRowxContext(ctx, QUERY_CHECK_ALIAS_EXISTS, req.Key, req.Value).Scan(&exists)
	if err != nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return exists, nil
}

func (c *client) ExistsInFile(ctx context.Context, req *ExistsInFileRequest) ([]string, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	if len(req.Aliases) == 0 {
		return nil, nil
	}

	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New("SELECT aliases.id FROM aliases").
		InnerJoin("files_aliases as fa", "fa.alias_id=aliases.id").
		InnerJoin("files", "fa.file_id=files.id").
		Where().IN("aliases.id", req.Aliases).EQ("files.id", req.FileID).Query()

	foundAliases := make([]string, 0)
	err := c.db.SelectContext(ctx, &foundAliases, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return foundAliases, nil
}

func (c *client) Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "alias_id", Value: req.AliasID},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return false, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New(QUERY_DELETE_ALIAS).Where().EQ("id", req.AliasID).Query()

	result, err := c.db.ExecContext(ctx, preparedQuery.String())
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

func (c *client) GetMany(ctx context.Context, req *GetManyRequest) ([]*Alias, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	if req.Limit == nil {
		req.Limit = utils.MakePointer(DEFAULT_LIMIT)
	}

	if req.Offset == nil {
		req.Offset = utils.MakePointer(DEFAULT_OFFSET)
	}

	preparedQuery := query.New(QUERY_GET_ALIASES).Limit(*req.Limit).Offset(*req.Offset)
	if req.OrderBy != nil && req.OrderType != nil {
		order := query.ASC
		if strings.ToUpper(*req.OrderType) == query.DESC || strings.ToUpper(*req.OrderType) == query.ASC {
			order = *req.OrderType
		}
		preparedQuery.Order(*req.OrderBy, order)
	} else {
		preparedQuery.Order("key", query.ASC)
	}

	if req.Key != nil {
		preparedQuery.Where().EQ("key", req.Key)
	}

	if req.Value != nil {
		preparedQuery.Where().EQ("value", req.Value)
	}

	aliases := make([]*Alias, 0)
	err := c.db.SelectContext(ctx, &aliases, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return aliases, nil
}

func (c *client) Get(ctx context.Context, req *GetRequest) (*Alias, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "alias_id", Value: req.AliasID},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New(QUERY_GET_ALIASES).Where().EQ("id", req.AliasID).Query()
	var alias Alias
	err := c.db.GetContext(ctx, &alias, preparedQuery.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.HTTPStatus(http.StatusNotFound))
		}
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &alias, nil
}

func (c *client) Edit(ctx context.Context, req *EditRequest) (*Alias, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "alias_id", Value: req.AliasID},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	if req.Color == nil && req.Key == nil && req.Value == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("key, value or color", "required"))
	}

	queryUpdate := query.New(QUERY_UPDATE_ALIAS).Set("updated_at", "now()").
		Where().EQ("id", req.AliasID).Query().
		Returning("id", "key", "value", "color", "created_at", "updated_at")

	if req.Color != nil {
		queryUpdate.Set("color", *req.Color)
	}

	if req.Key != nil {
		queryUpdate.Set("key", *req.Key)
	}

	if req.Value != nil {
		queryUpdate.Set("value", *req.Value)
	}

	var alias Alias
	err := c.db.QueryRowxContext(ctx, queryUpdate.String()).StructScan(&alias)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &alias, nil
}

func (c *client) AddToFile(ctx context.Context, req *AddToFileRequest) (int, tiny_errors.ErrorHandler) {
	if req == nil {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	if len(req.Aliases) == 0 {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("aliases", "required"))
	}

	preparedQuery := query.New(QUERY_ADD_TO_FILE).InsertColumns("file_id", "alias_id")

	for _, aliasID := range req.Aliases {
		preparedQuery.Values(req.FileID, aliasID)
	}

	result, err := c.db.ExecContext(ctx, preparedQuery.String())
	if err != nil {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return int(affected), nil
}

func (c *client) RemoveFromFile(ctx context.Context, req *RemoveFromFileRequest) (int, tiny_errors.ErrorHandler) {
	if req == nil {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	if len(req.Aliases) == 0 {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, tiny_errors.Detail("aliases", "required"))
	}

	preparedQuery := query.New(QUERY_REMOVE_FROM_FILE).Where().EQ("file_id", req.FileID).IN("alias_id", req.Aliases).Query()

	result, err := c.db.ExecContext(ctx, preparedQuery.String())
	if err != nil {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return int(affected), nil
}

func (c *client) GetFileAliases(ctx context.Context, req *GetFileAliasesRequest) ([]*Alias, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileID},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New("SELECT a.id, a.key, a.value, a.color, a.created_at, a.updated_at FROM aliases as a").
		InnerJoin("files_aliases as fa", "fa.alias_id=a.id").
		Where().EQ("fa.file_id", req.FileID).Query()
	aliases := make([]*Alias, 0)
	err := c.db.SelectContext(ctx, &aliases, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return aliases, nil
}

func (c *client) GetFilesAliasesManyToMany(ctx context.Context, req *GetFilesAliasesManyToManyRequest) ([]*AliasWithFileId, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	requiredFields := []utils.RequiredField{
		{Name: "file_id", Value: req.FileIDs},
	}
	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	preparedQuery := query.New("SELECT a.id, fa.file_id, a.key, a.value, a.color, a.created_at, a.updated_at FROM aliases as a").
		InnerJoin("files_aliases as fa", "fa.alias_id=a.id").
		Where().IN("fa.file_id", req.FileIDs).Query()
	aliases := make([]*AliasWithFileId, 0)
	err := c.db.SelectContext(ctx, &aliases, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return aliases, nil
}
