package models

import (
	"github.com/Moranilt/config-keeper/pkg/file_contents"
	"github.com/Moranilt/config-keeper/pkg/files"
	"github.com/Moranilt/config-keeper/pkg/folders"
	"github.com/Moranilt/config-keeper/pkg/listeners"
)

type CreateFolderRequest struct {
	Name     string  `json:"name"`
	ParentID *string `json:"parent_id"`
}

type CreateFolderResponse folders.Folder

type GetFolderRequest struct {
	ID          string  `mapstructure:"id"`
	OrderColumn *string `mapstructure:"order_column"`
	OrderType   *string `mapstructure:"order_type"`
}

type GetFolderResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	ParentID  *string           `json:"parent_id"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
	Path      string            `json:"path"`
	Folders   []*folders.Folder `json:"folders"`
	Files     []*files.File     `json:"files"`
}

type DeleteFolderRequest struct {
	ID string `mapstructure:"id"`
}

type DeleteFolderResponse struct {
	Status bool `json:"status"`
}

type EditFolderRequest struct {
	ID   string `mapstructure:"id"`
	Name string `json:"name"`
}

type EditFolderResponse folders.Folder

type CreateFileRequest struct {
	Name     string  `json:"name"`
	FolderID *string `json:"folder_id"`
}

type CreateFileResponse files.File

type DeleteFileRequest struct {
	ID string `mapstructure:"id"`
}

type DeleteFileResponse struct {
	Status bool `json:"status"`
}

type EditFileRequest struct {
	ID   string `mapstructure:"id"`
	Name string `json:"name"`
}

type EditFileResponse files.File

type GetFileRequest struct {
	ID string `mapstructure:"id"`
}

type GetFileResponse struct {
	files.File
	Contents []*file_contents.FileContent `json:"contents"`
}

type CreateFileContentRequest struct {
	FileID  string `mapstructure:"file_id"`
	Version string `json:"version"`
	Content string `json:"content"`
}

type CreateFileContentResponse file_contents.FileContent

type GetFileContentsRequest struct {
	FileID  string  `mapstructure:"file_id"`
	Version *string `mapstructure:"version"`
}

type GetFileContentsResponse []*file_contents.FileContent

type EditFileContentRequest struct {
	ID      string  `mapstructure:"id"`
	Version *string `json:"version"`
	Content *string `json:"content"`
}

type EditFileContentResponse file_contents.FileContent

type DeleteFileContentRequest struct {
	ID string `mapstructure:"id"`
}

type DeleteFileContentResponse struct {
	Status bool `json:"status"`
}

type CreateListenerRequest struct {
	FileID           string `mapstructure:"file_id"`
	Name             string `json:"name"`
	CallbackEndpoint string `json:"callback_endpoint"`
}

type CreateListenerResponse listeners.Listener

type GetListenerRequest struct {
	ID string `mapstructure:"id"`
}

type GetListenerResponse listeners.Listener

type GetFileListenersRequest struct {
	FileID string `mapstructure:"file_id"`
}

type GetFileListenersResponse []*listeners.Listener

type EditListenerRequest struct {
	ListenerID       string  `mapstructure:"id"`
	Name             *string `json:"name"`
	CallbackEndpoint *string `json:"callback_endpoint"`
}

type EditListenerResponse listeners.Listener

type DeleteListenerRequest struct {
	ID string `mapstructure:"id"`
}

type DeleteListenerResponse struct {
	Status bool `json:"status"`
}
