package folders

const (
	QUERY_INSERT_FOLDER                  = "INSERT INTO folders (name, parent_id) VALUES($1, $2) RETURNING id, name, parent_id, created_at, updated_at"
	QUERY_DEFAULT_SELECT_FOLDERS_ID      = "SELECT id FROM folders"
	QUERY_GET_FOLDER_WITH_PARENT_ID_NULL = "SELECT id FROM folders WHERE name = $1 AND parent_id IS NULL"
	QUERY_GET_FOLDER_WITH_PARENT_ID      = "SELECT id FROM folders WHERE name = $1 AND parent_id = $2"
	QUERY_GET_FOLDER_WITH_PATH           = `WITH RECURSIVE folder_path AS (
		SELECT 
			id,
			parent_id,
			name,
			CAST(name AS VARCHAR) AS path,
			created_at,
			updated_at
		FROM 
			folders
		WHERE 
			parent_id IS NULL
		
		UNION ALL
		SELECT 
			f.id,
			f.parent_id,
			f.name,
			CONCAT(fp.path, '/', f.name) AS path,
			f.created_at,
			f.updated_at
		FROM 
			folders f
			JOIN folder_path fp ON f.parent_id = fp.id
	)
	SELECT 
		id,
		parent_id,
		name,
		path,
		created_at,
		updated_at
	FROM 
		folder_path`
	QUERY_GET_FOLDERS                               = "SELECT id, name, parent_id, created_at, updated_at FROM folders"
	QUERY_DELETE_FOLDER                             = "DELETE FROM folders WHERE id = $1"
	QUERY_UPDATE_FOLDER                             = "UPDATE folders SET name = $1, updated_at = now() WHERE id = $2 RETURNING id, name, parent_id, created_at, updated_at"
	QUERY_CHECK_FOLDER_EXISTS_BY_PARENT_ID_AND_NAME = `SELECT EXISTS(
	SELECT 1 FROM folders 
	WHERE name = $1
	AND (
			(parent_id IS NULL AND (SELECT parent_id FROM folders WHERE id = $2) is NULL)
			OR
			(parent_id = (SELECT parent_id FROM folders WHERE id = $2))
	))`
)

type CreateRequest struct {
	Name     string
	ParentID *string
}

type Folder struct {
	ID        string  `json:"id" db:"id"`
	Name      string  `json:"name" db:"name"`
	ParentID  *string `json:"parent_id" db:"parent_id"`
	CreatedAt string  `json:"created_at" db:"created_at"`
	UpdatedAt string  `json:"updated_at" db:"updated_at"`
}

type FolderWithPath struct {
	Folder
	Path string `json:"path" db:"path"`
}

type ExistsRequest struct {
	Name     *string
	ParentID *string
}

type GetRequest struct {
	ID string
}

type Order struct {
	Column *string
	Type   *string
}

type GetManyRequest struct {
	ParentID *string
	Order    *Order
}

type DeleteRequest struct {
	ID string
}

type EditRequest struct {
	ID   string
	Name string
}
