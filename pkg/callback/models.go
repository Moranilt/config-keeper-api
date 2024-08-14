package callback

import (
	"github.com/Moranilt/config-keeper/pkg/file_contents"
	"github.com/Moranilt/config-keeper/pkg/files"
)

type CallbackRequest struct {
	FileID string
}

type FileData struct {
	files.File
	FileContent []*file_contents.FileContent `json:"file_contents"`
}
