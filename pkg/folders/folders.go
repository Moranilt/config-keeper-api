package folders

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/query"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/jmoiron/sqlx"
)

type client struct {
	db *database.Client
}

type Client interface {
	New(ctx context.Context, req *NewRequest) (*Folder, tiny_errors.ErrorHandler)
	Exists(ctx context.Context, req *ExistsRequest) (bool, tiny_errors.ErrorHandler)
	Get(ctx context.Context, req *GetRequest) (*FolderWithPath, tiny_errors.ErrorHandler)
	GetMany(ctx context.Context, req *GetManyRequest) ([]*Folder, tiny_errors.ErrorHandler)
	Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler)
	Edit(ctx context.Context, req *EditRequest) (*Folder, tiny_errors.ErrorHandler)
}

func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

func (c *client) New(ctx context.Context, req *NewRequest) (*Folder, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	row := c.db.QueryRowxContext(ctx, QUERY_INSERT_FOLDER, req.Name, req.ParentID)
	if row.Err() != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var folder Folder
	err := row.StructScan(&folder)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Marshal, tiny_errors.Message(err.Error()))
	}

	return &folder, nil
}

func (c *client) Exists(ctx context.Context, req *ExistsRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	var (
		row *sqlx.Row
		id  string
	)
	preparedQuery := query.New(QUERY_DEFAULT_SELECT_FOLDERS_ID)
	if req.ParentID == nil {
		preparedQuery.Where().IS("parent_id", nil)
	} else {
		preparedQuery.Where().EQ("parent_id", req.ParentID)
	}

	if req.Name != nil {
		preparedQuery.Where().EQ("name", req.Name)
	}

	row = c.db.QueryRowxContext(ctx, preparedQuery.String())

	if row.Err() != nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, tiny_errors.New(custom_errors.ERR_CODE_Marshal, tiny_errors.Message(err.Error()))
	}

	return true, nil
}

func (c *client) Get(ctx context.Context, req *GetRequest) (*FolderWithPath, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_FOLDER_WITH_PATH)
	preparedQuery.Where().EQ("id", req.ID)

	var folder FolderWithPath
	err := c.db.GetContext(ctx, &folder, preparedQuery.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.HTTPStatus(http.StatusNotFound))
		}
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return &folder, nil
}

func (c *client) GetMany(ctx context.Context, req *GetManyRequest) ([]*Folder, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	preparedQuery := query.New(QUERY_GET_FOLDERS)
	if req.Order != nil && req.Order.Column != nil && req.Order.Type != nil {
		order := query.ASC
		orderType := strings.ToUpper(*req.Order.Type)
		if orderType == query.DESC || orderType == query.ASC {
			order = orderType
		}
		preparedQuery = preparedQuery.Order(*req.Order.Column, order)
	} else {
		preparedQuery = preparedQuery.Order("name", query.ASC)
	}

	if req.ParentID == nil {
		preparedQuery.Where().IS("parent_id", nil)
	} else {
		preparedQuery.Where().EQ("parent_id", req.ParentID)
	}

	folders := make([]*Folder, 0)
	err := c.db.SelectContext(ctx, &folders, preparedQuery.String())
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return folders, nil
}

func (c *client) Delete(ctx context.Context, req *DeleteRequest) (bool, tiny_errors.ErrorHandler) {
	if req == nil {
		return false, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	result, err := c.db.ExecContext(ctx, QUERY_DELETE_FOLDER, req.ID)
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

func (c *client) Edit(ctx context.Context, req *EditRequest) (*Folder, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	if req.ID == "" || req.Name == "" {
		return nil, tiny_errors.New(
			custom_errors.ERR_CODE_NotValid,
			tiny_errors.Detail("id", "required"),
			tiny_errors.Detail("name", "required"),
		)
	}

	row := c.db.QueryRowxContext(
		ctx,
		QUERY_CHECK_FOLDER_EXISTS_BY_PARENT_ID_AND_NAME,
		req.Name,
		req.ID,
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
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("folder with such name already exists"))
	}

	row = c.db.QueryRowxContext(ctx, QUERY_UPDATE_FOLDER, req.Name, req.ID)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			return nil, tiny_errors.New(custom_errors.ERR_CODE_NotFound, tiny_errors.HTTPStatus(http.StatusNotFound))
		}
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(row.Err().Error()))
	}

	var folder Folder
	err = row.StructScan(&folder)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Marshal, tiny_errors.Message(err.Error()))
	}

	return &folder, nil
}
