package listeners

const (
	QUERY_CREATE_LISTENER = "INSERT INTO listeners (file_id, callback_endpoint, name) VALUES ($1, $2, $3) RETURNING id, file_id, callback_endpoint, name, created_at, updated_at"
	QUERY_GET_LISTENERS   = "SELECT id, file_id, callback_endpoint, name, created_at, updated_at FROM listeners"
	QUERY_DELETE_LISTENER = "DELETE FROM listeners WHERE id = $1"
)

type Listener struct {
	ID               string `db:"id" json:"id"`
	FileID           string `db:"file_id" json:"file_id"`
	CallbackEndpoint string `db:"callback_endpoint" json:"callback_endpoint"`
	Name             string `db:"name" json:"name"`
	CreatedAt        string `db:"created_at" json:"created_at"`
	UpdatedAt        string `db:"updated_at" json:"updated_at"`
}

type CreateRequest struct {
	Name             string `json:"name"`
	CallbackEndpoint string `json:"callback_endpoint"`
	FileID           string `json:"file_id"`
}

type GetManyRequest struct {
	FileID string `json:"file_id"`
}

type GetRequest struct {
	ID string `json:"id"`
}

type DeleteRequest struct {
	ID string `json:"id"`
}

type EditRequest struct {
	ID               string  `json:"id"`
	Name             *string `json:"name"`
	CallbackEndpoint *string `json:"callback_endpoint"`
}
