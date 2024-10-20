DROP INDEX IF EXISTS idx_mv_file_search_trigram;
DROP EXTENSION IF EXISTS pg_trgm;
DROP INDEX IF EXISTS mv_file_search_id_idx;

DROP TRIGGER IF EXISTS refresh_mv_file_search_folders ON folders;
DROP TRIGGER IF EXISTS refresh_mv_file_search_files_aliases ON files_aliases;
DROP TRIGGER IF EXISTS refresh_mv_file_search_aliases ON aliases;
DROP TRIGGER IF EXISTS refresh_mv_file_search_files ON files;

DROP FUNCTION IF EXISTS refresh_mv_file_search();

DROP MATERIALIZED VIEW IF EXISTS mv_file_search;

DROP INDEX IF EXISTS idx_folders_name_fts;
DROP INDEX IF EXISTS idx_aliases_value_fts;
DROP INDEX IF EXISTS idx_aliases_key_fts;
DROP INDEX IF EXISTS idx_files_name_fts;