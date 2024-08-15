package file_contents

const (
	QUERY_CREATE_CONTENT = `WITH inserted_row AS (
    INSERT INTO file_contents (file_id, version, content, format_id)
    VALUES ($1, $2, $3, $4)
    RETURNING id, file_id, version, format_id, content, created_at, updated_at
	)
	SELECT 
			i.id, 
			i.file_id, 
			i.version,
			i.content, 
			i.created_at, 
			i.updated_at,
			cf.name AS format
	FROM inserted_row i
	LEFT JOIN content_formats cf ON i.format_id = cf.id`
	QUERY_GET_FILES_CONTENT_ID_BY_VERSION = "SELECT id FROM file_contents WHERE file_id = $1 AND version = $2"
	QUERY_GET_FILE_CONTENTS               = `SELECT fc.id, file_id, cf.name AS format, version, content, created_at, updated_at 
	FROM file_contents AS fc 
	LEFT JOIN content_formats AS cf ON cf.id = fc.format_id`
	QUERY_GET_FILE_CONTENTS_ID = "SELECT id FROM file_contents"
	QUERY_DELETE_FILE_CONTENT  = "DELETE FROM file_contents WHERE id = $1"
)

type FileContent struct {
	ID        string `json:"id" db:"id"`
	Content   string `json:"content" db:"content"`
	Version   string `json:"version" db:"version"`
	FileID    string `json:"file_id" db:"file_id"`
	Format    string `json:"format" db:"format"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type CreateRequest struct {
	FileID   string
	Version  string
	Content  string
	FormatID string
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
