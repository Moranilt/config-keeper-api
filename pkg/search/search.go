package search

import (
	"context"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/tiny_errors"
)

type client struct {
	db *database.Client
}

type Client interface {
	// Default global search using materialized view
	Global(ctx context.Context, req *GlobalSearchRequest) ([]*GlobalSearchResult, tiny_errors.ErrorHandler)
}

// New creates a new instance of the Client interface, which provides methods for
// interacting with search in a database.
func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

func (c *client) Global(ctx context.Context, req *GlobalSearchRequest) ([]*GlobalSearchResult, tiny_errors.ErrorHandler) {
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}

	requiredFields := []utils.RequiredField{
		{Name: "query", Value: req.Query},
	}

	requiredErr := utils.ValidateRequiredFields(requiredFields)
	if len(requiredErr) > 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, requiredErr...)
	}

	result := make([]*GlobalSearchResult, 0)
	err := c.db.SelectContext(ctx, &result, QUERY_GLOBAL_SEARCH, req.Query)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return result, nil
}
