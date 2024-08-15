CREATE TABLE content_formats (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  name VARCHAR(255) NOT NULL
);

INSERT INTO content_formats (name) VALUES ('yaml'), ('toml'), ('json'), ('env');

CREATE TABLE file_contents (
  id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
  file_id UUID NOT NULL,
  content TEXT NOT NULL,
  version VARCHAR(255) NOT NULL,
  format_id UUID NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
  FOREIGN KEY (format_id) REFERENCES content_formats(id) ON DELETE CASCADE
);