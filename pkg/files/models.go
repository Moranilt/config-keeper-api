package files

const (
	QUERY_GET_FILES                               = "SELECT id, folder_id, name, created_at, updated_at FROM files"
	QUERY_CREATE_FILE                             = "INSERT INTO files (folder_id, name) VALUES ($1, $2) RETURNING id, folder_id, name, created_at, updated_at"
	QUERY_FILE_EXISTS                             = "SELECT EXISTS(SELECT 1 FROM files WHERE id=$1)"
	QUERY_DELETE_FILE                             = "DELETE FROM files WHERE id=$1"
	QUERY_CHECK_FILE_EXISTS_BY_FOLDER_ID_AND_NAME = `SELECT EXISTS(
		SELECT 1 FROM files 
		WHERE name = $1
		AND (
				(folder_id IS NULL AND (SELECT folder_id FROM files WHERE id = $2) is NULL)
				OR
				(folder_id = (SELECT folder_id FROM files WHERE id = $2))
		))`
	QUERY_UPDATE_FILE = "UPDATE files SET name = $1, updated_at = now() WHERE id = $2 RETURNING id, folder_id, name, created_at, updated_at"
)

type File struct {
	ID        string  `json:"id" db:"id"`
	FolderID  *string `json:"folder_id" db:"folder_id"`
	Name      string  `json:"name" db:"name"`
	CreatedAt string  `json:"created_at" db:"created_at"`
	UpdatedAt string  `json:"updated_at" db:"updated_at"`
}

type Order struct {
	Column *string
	Type   *string
}

type GetFilesInFolderRequest struct {
	FolderID *string
	Order    *Order
}

type CreateRequest struct {
	FolderID *string
	Name     string
}

type DeleteRequest struct {
	ID string
}

type EditRequest struct {
	Name   string
	FileID string
}

type GetRequest struct {
	ID string
}
