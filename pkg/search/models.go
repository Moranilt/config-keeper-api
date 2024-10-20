package search

const (
	QUERY_GLOBAL_SEARCH = `WITH search_query AS (
    SELECT plainto_tsquery('english', $1) AS tsquery,
		$1 as raw_query
)
SELECT DISTINCT ON (mv.id) 
    mv.id, mv.folder_name, mv.file_name, mv.folder_id, mv.aliases, mv.created_at, mv.updated_at
FROM 
    mv_file_search mv,
    search_query
WHERE 
    mv.search_vector @@ search_query.tsquery
		OR mv.file_name ILIKE '%' || search_query.raw_query || '%'
    OR mv.aliases ILIKE '%' || search_query.raw_query || '%'
    OR mv.folder_name ILIKE '%' || search_query.raw_query || '%'
ORDER BY 
    mv.id, 
    ts_rank(mv.search_vector, search_query.tsquery) DESC,
		similarity(mv.file_name, search_query.raw_query) DESC,
    similarity(mv.aliases, search_query.raw_query) DESC,
    similarity(mv.folder_name, search_query.raw_query) DESC`
)

type GlobalSearchRequest struct {
	Query string `json:"query"`
}

type GlobalSearchResult struct {
	ID         string  `json:"id" db:"id"`
	FolderName string  `json:"folder_name" db:"folder_name"`
	FileName   string  `json:"file_name" db:"file_name"`
	FolderID   *string `json:"folder_id" db:"folder_id"`
	Aliases    string  `json:"aliases" db:"aliases"`
	CreatedAt  string  `json:"created_at" db:"created_at"`
	UpdatedAt  string  `json:"updated_at" db:"updated_at"`
}
