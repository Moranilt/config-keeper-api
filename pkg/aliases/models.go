package aliases

const (
	QUERY_GET_ALIASES        = "SELECT * FROM aliases"
	QUERY_CREATE_ALIAS       = "INSERT INTO aliases"
	QUERY_CHECK_ALIAS_EXISTS = `SELECT EXISTS(
		SELECT 1 FROM folders 
		WHERE name = $1 and value = $2
		)`
	QUERY_CHECK_ALIASES_IN_FILE = `SELECT id FROM aliases INNER JOIN`
	QUERY_DELETE_ALIAS          = "DELETE FROM aliases"
	QUERY_UPDATE_ALIAS          = "UPDATE aliases"
	QUERY_ADD_TO_FILE           = "INSERT INTO files_aliases"
	QUERY_REMOVE_FROM_FILE      = "DELETE FROM files_aliases"
)

type Alias struct {
	ID        string `json:"id" db:"id"`
	Key       string `json:"key" db:"key"`
	Value     string `json:"value" db:"value"`
	Color     string `json:"color" db:"color"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type CreateRequest struct {
	Key   string
	Value string
	Color string
}

type ExistsRequest struct {
	Key   string
	Value string
}

type ExistsInFileRequest struct {
	FileID  string
	Aliases []string
}

type DeleteRequest struct {
	AliasID string
}

type GetManyRequest struct {
	Limit     *string
	Offset    *string
	Key       *string
	Value     *string
	OrderBy   *string
	OrderType *string
}

type EditRequest struct {
	AliasID string
	Key     *string
	Value   *string
	Color   *string
}

type AddToFileRequest struct {
	FileID  string
	Aliases []string
}

type RemoveFromFileRequest struct {
	FileID  string
	Aliases []string
}
