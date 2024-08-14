package file_contents

const (
	QUERY_CREATE_CONTENT                  = "INSERT INTO file_contents (file_id, version, content) VALUES ($1, $2, $3) RETURNING id, file_id, version, content, created_at, updated_at"
	QUERY_GET_FILES_CONTENT_ID_BY_VERSION = "SELECT id FROM file_contents WHERE file_id = $1 AND version = $2"
	QUERY_GET_FILE_CONTENTS               = "SELECT id, file_id, version, content, created_at, updated_at FROM file_contents"
	QUERY_GET_FILE_CONTENTS_ID            = "SELECT id FROM file_contents"
	QUERY_DELETE_FILE_CONTENT             = "DELETE FROM file_contents WHERE id = $1"
)

type FileContent struct {
	ID        string `json:"id" db:"id"`
	Content   string `json:"content" db:"content"`
	Version   string `json:"version" db:"version"`
	FileID    string `json:"file_id" db:"file_id"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type CreateRequest struct {
	FileID  string
	Version string
	Content string
}

type GetManyRequest struct {
	FileID  string
	Version *string
}

type EditRequest struct {
	FileContentID string
	Content       *string
	Version       *string
}

type DeleteRequest struct {
	ID string
}
