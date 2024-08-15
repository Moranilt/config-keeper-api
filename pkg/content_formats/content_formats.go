package content_formats

import (
	"context"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/tiny_errors"
)

type client struct {
	db *database.Client
}

type Client interface {
	// GetMany retrieves multiple content formats entries from the database.
	GetMany(ctx context.Context) ([]*ContentFormat, tiny_errors.ErrorHandler)
}

// New creates a new instance of the Client interface, which provides methods for
// interacting with content formats in a database.
func New(db *database.Client) Client {
	return &client{
		db: db,
	}
}

func (c *client) GetMany(ctx context.Context) ([]*ContentFormat, tiny_errors.ErrorHandler) {
	var contentFormats []*ContentFormat
	err := c.db.SelectContext(ctx, &contentFormats, QUERY_GET_FORMATS)
	if err != nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Database, tiny_errors.Message(err.Error()))
	}

	return contentFormats, nil
}
