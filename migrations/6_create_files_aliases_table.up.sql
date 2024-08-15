CREATE TABLE files_aliases (
  alias_id UUID NOT NULL,
  file_id UUID NOT NULL,
  FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
  FOREIGN KEY (alias_id) REFERENCES aliases(id) ON DELETE CASCADE
);