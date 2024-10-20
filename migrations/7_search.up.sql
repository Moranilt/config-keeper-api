CREATE INDEX idx_files_name_fts ON files USING gin(to_tsvector('english', name));
CREATE INDEX idx_aliases_key_fts ON aliases USING gin(to_tsvector('english', key));
CREATE INDEX idx_aliases_value_fts ON aliases USING gin(to_tsvector('english', value));
CREATE INDEX idx_folders_name_fts ON folders USING gin(to_tsvector('english', name));

CREATE MATERIALIZED VIEW mv_file_search AS
WITH split_names AS (
    SELECT 
        f.id,
        regexp_split_to_table(f.name, '[_\-\.]') AS name_part
    FROM 
        files f
)
SELECT 
    f.id,
    f.folder_id,
    f.created_at,
    f.updated_at,
    string_agg(DISTINCT f.name, ' ') AS file_name,
    CASE 
        WHEN COUNT(a.id) > 0 THEN string_agg(DISTINCT CONCAT(a.key, '=', a.value), ';')
        ELSE NULL
    END AS aliases,
    string_agg(DISTINCT COALESCE(folders.name, ''), ' ') AS folder_name,
    to_tsvector('english', string_agg(DISTINCT f.name, ' ')) ||
    to_tsvector('english', string_agg(DISTINCT sn.name_part, ' ')) ||
    to_tsvector('english', string_agg(DISTINCT COALESCE(folders.name, ''), ' ')) ||
    to_tsvector('english', COALESCE(string_agg(DISTINCT CONCAT(a.key, '=', a.value), ' '), '')) ||
    to_tsvector('simple', string_agg(DISTINCT f.name, ' ')) ||
    to_tsvector('simple', string_agg(DISTINCT sn.name_part, ' ')) ||
    to_tsvector('simple', string_agg(DISTINCT COALESCE(folders.name, ''), ' ')) ||
    to_tsvector('simple', COALESCE(string_agg(DISTINCT CONCAT(a.key, '=', a.value), ' '), '')) AS search_vector
FROM 
    files f
LEFT JOIN 
    split_names sn ON f.id = sn.id
LEFT JOIN 
    files_aliases fa ON f.id = fa.file_id
LEFT JOIN 
    aliases a ON fa.alias_id = a.id
LEFT JOIN 
    folders ON folders.id = f.folder_id
GROUP BY 
    f.id;


CREATE OR REPLACE FUNCTION refresh_mv_file_search()
RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_file_search;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER refresh_mv_file_search_files
AFTER INSERT OR UPDATE OR DELETE ON files
FOR EACH STATEMENT EXECUTE FUNCTION refresh_mv_file_search();

CREATE TRIGGER refresh_mv_file_search_aliases
AFTER INSERT OR UPDATE OR DELETE ON aliases
FOR EACH STATEMENT EXECUTE FUNCTION refresh_mv_file_search();

CREATE TRIGGER refresh_mv_file_search_files_aliases
AFTER INSERT OR UPDATE OR DELETE ON files_aliases
FOR EACH STATEMENT EXECUTE FUNCTION refresh_mv_file_search();

CREATE TRIGGER refresh_mv_file_search_folders
AFTER INSERT OR UPDATE OR DELETE ON folders
FOR EACH STATEMENT EXECUTE FUNCTION refresh_mv_file_search();

CREATE UNIQUE INDEX ON mv_file_search (id);
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_mv_file_search_trigram ON mv_file_search USING gin (
    file_name gin_trgm_ops,
    aliases gin_trgm_ops,
    folder_name gin_trgm_ops
);

-- WITH search_query AS (
--     SELECT plainto_tsquery('english', 'folder_2') AS tsquery,
-- 	'folder_2' as raw_query
-- )
-- SELECT DISTINCT ON (mv.id) 
--     mv.id, mv.folder_name, mv.file_name, mv.folder_id, mv.aliases, mv.created_at, mv.updated_at
-- FROM 
--     mv_file_search mv,
--     search_query
-- WHERE 
--     mv.search_vector @@ search_query.tsquery
-- 	OR mv.file_name ILIKE '%' || search_query.raw_query || '%'
--     OR mv.aliases ILIKE '%' || search_query.raw_query || '%'
--     OR mv.folder_name ILIKE '%' || search_query.raw_query || '%'
-- ORDER BY 
--     mv.id, 
--     ts_rank(mv.search_vector, search_query.tsquery) DESC,
-- 	similarity(mv.file_name, search_query.raw_query) DESC,
--     similarity(mv.aliases, search_query.raw_query) DESC,
--     similarity(mv.folder_name, search_query.raw_query) DESC

