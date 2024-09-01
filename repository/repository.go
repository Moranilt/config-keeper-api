package repository

import (
	"context"
	"slices"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/models"
	"github.com/Moranilt/config-keeper/pkg/aliases"
	"github.com/Moranilt/config-keeper/pkg/callback"
	"github.com/Moranilt/config-keeper/pkg/content_formats"
	"github.com/Moranilt/config-keeper/pkg/file_contents"
	"github.com/Moranilt/config-keeper/pkg/files"
	"github.com/Moranilt/config-keeper/pkg/folders"
	"github.com/Moranilt/config-keeper/pkg/listeners"
	"github.com/Moranilt/config-keeper/utils"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/logger"
	"github.com/Moranilt/http-utils/tiny_errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const TracerName string = "repository"

type Repository struct {
	db             *database.Client
	log            logger.Logger
	tracer         trace.Tracer
	folders        folders.Client
	files          files.Client
	fileContent    file_contents.Client
	listeners      listeners.Client
	callback       callback.CallbackChannel
	contentFormats content_formats.Client
	aliases        aliases.Client
}

func New(
	db *database.Client,
	callback callback.CallbackChannel,
	folders folders.Client,
	files files.Client,
	fileContent file_contents.Client,
	listeners listeners.Client,
	contentFormats content_formats.Client,
	aliases aliases.Client,
	logger logger.Logger,
) *Repository {
	return &Repository{
		db:             db,
		log:            logger,
		tracer:         otel.Tracer(TracerName),
		folders:        folders,
		files:          files,
		fileContent:    fileContent,
		listeners:      listeners,
		callback:       callback,
		contentFormats: contentFormats,
		aliases:        aliases,
	}
}

// CreateFolder creates a new folder in the repository. If a folder with the same name and parent already exists, it returns an error.
//
// The function logs the request, starts a new trace span, clears the folder name, checks if the folder already exists, and creates a new folder if it doesn't. It returns the created folder's details in the response.
func (repo *Repository) CreateFolder(ctx context.Context, req *models.CreateFolderRequest) (*models.CreateFolderResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	ctx, span := repo.tracer.Start(ctx, "CreateFolder", trace.WithAttributes(
		attribute.String("name", req.Name),
	))
	defer span.End()

	clearName, err := utils.ClearName(req.Name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ClearName")
		return nil, err
	}

	exists, err := repo.folders.Exists(ctx, &folders.ExistsRequest{
		Name:     &clearName,
		ParentID: req.ParentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ExistsFolder")
		return nil, err
	}

	if exists {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists)
	}

	folder, err := repo.folders.Create(ctx, &folders.CreateRequest{
		Name:     clearName,
		ParentID: req.ParentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "NewFolder")
		return nil, err
	}

	return &models.CreateFolderResponse{
		ID:        folder.ID,
		Name:      folder.Name,
		ParentID:  folder.ParentID,
		CreatedAt: folder.CreatedAt,
		UpdatedAt: folder.UpdatedAt,
	}, nil
}

func (repo *Repository) GetFolder(ctx context.Context, req *models.GetFolderRequest) (*models.GetFolderResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "GetFolder", trace.WithAttributes(
		attribute.String("folder_id", req.FolderID),
	))
	defer span.End()

	var (
		folderWithPath *folders.FolderWithPath
		err            tiny_errors.ErrorHandler
		parentID       *string
	)

	if req.FolderID == "root" {
		folderWithPath = &folders.FolderWithPath{
			Folder: folders.Folder{
				ID:        "root",
				Name:      "root",
				ParentID:  nil,
				CreatedAt: "1979-01-01T00:00:00Z",
				UpdatedAt: "1979-01-01T00:00:00Z",
			},
			Path: "root",
		}
	} else {
		parentID = &req.FolderID
		folderWithPath, err = repo.folders.Get(ctx, &folders.GetRequest{
			ID: req.FolderID,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "GetFolder")
			return nil, err
		}
	}

	folders, err := repo.folders.GetMany(ctx, &folders.GetManyRequest{
		ParentID: parentID,
		Order: &folders.Order{
			Column: req.OrderColumn,
			Type:   req.OrderType,
		},
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFolders")
		return nil, err
	}

	files, err := repo.files.GetMany(ctx, &files.GetManyRequest{
		FolderID: parentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFiles")
		return nil, err
	}

	fileIds := make([]string, len(files))
	for i, file := range files {
		fileIds[i] = file.ID
	}

	foundAliases, err := repo.aliases.GetFilesAliasesManyToMany(ctx, &aliases.GetFilesAliasesManyToManyRequest{
		FileIDs: fileIds,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFilesAliasesManyToMany")
		return nil, err
	}

	fileAliases := make(map[string][]*aliases.Alias)
	for _, alias := range foundAliases {
		fileAliases[alias.FileID] = append(fileAliases[alias.FileID], &aliases.Alias{
			ID:        alias.ID,
			Key:       alias.Key,
			Value:     alias.Value,
			Color:     alias.Color,
			CreatedAt: alias.CreatedAt,
			UpdatedAt: alias.UpdatedAt,
		})
	}

	filesWithAliases := make([]*models.FileWithAliases, len(files))
	for i, file := range files {
		fileWithAliases := &models.FileWithAliases{
			File:    *file,
			Aliases: fileAliases[file.ID],
		}
		filesWithAliases[i] = fileWithAliases
	}

	return &models.GetFolderResponse{
		ID:        folderWithPath.ID,
		Name:      folderWithPath.Name,
		ParentID:  folderWithPath.ParentID,
		CreatedAt: folderWithPath.CreatedAt,
		UpdatedAt: folderWithPath.UpdatedAt,
		Path:      folderWithPath.Path,
		Folders:   folders,
		Files:     filesWithAliases,
	}, nil
}

func (repo *Repository) DeleteFolder(ctx context.Context, req *models.DeleteFolderRequest) (*models.DeleteFolderResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "DeleteFolder", trace.WithAttributes(
		attribute.String("folder_id", req.FolderID),
	))
	defer span.End()

	removed, err := repo.folders.Delete(ctx, &folders.DeleteRequest{
		ID: req.FolderID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete")
		return nil, err
	}

	return &models.DeleteFolderResponse{
		Status: removed,
	}, nil
}

func (repo *Repository) EditFolder(ctx context.Context, req *models.EditFolderRequest) (*models.EditFolderResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "EditFolder", trace.WithAttributes(
		attribute.String("folder_id", req.FolderID),
		attribute.String("name", req.Name),
	))
	defer span.End()

	clearName, err := utils.ClearName(req.Name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ClearName")
		return nil, err
	}

	requiredFields := []utils.RequiredField{
		{
			Name:  "id",
			Value: req.FolderID,
		},
		{
			Name:  "name",
			Value: clearName,
		},
	}

	errFields := utils.ValidateRequiredFields(requiredFields)
	if errFields != nil {
		err := tiny_errors.New(custom_errors.ERR_CODE_REQUIRED_FIELD, errFields...)
		span.RecordError(err)
		span.SetStatus(codes.Error, "ValidateRequiredFields")
		return nil, err
	}

	folder, err := repo.folders.Edit(ctx, &folders.EditRequest{
		ID:   req.FolderID,
		Name: clearName,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "EditFolder")
		return nil, err
	}

	return (*models.EditFolderResponse)(folder), nil
}

func (repo *Repository) CreateFile(ctx context.Context, req *models.CreateFileRequest) (*models.CreateFileResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	ctx, span := repo.tracer.Start(ctx, "CreateFile", trace.WithAttributes(
		attribute.String("name", req.Name),
	))
	defer span.End()

	clearName, err := utils.ClearName(req.Name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ClearName")
		return nil, err
	}

	exists, err := repo.folders.Exists(ctx, &folders.ExistsRequest{
		Name:     &clearName,
		ParentID: req.FolderID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ExistsFolder")
		return nil, err
	}

	if exists {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists)
	}

	file, err := repo.files.Create(ctx, &files.CreateRequest{
		Name:     clearName,
		FolderID: req.FolderID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "CreateFile")
		return nil, err
	}

	return (*models.CreateFileResponse)(file), nil
}

func (repo *Repository) DeleteFile(ctx context.Context, req *models.DeleteFileRequest) (*models.DeleteFileResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "DeleteFile", trace.WithAttributes(
		attribute.String("id", req.ID),
	))
	defer span.End()

	removed, err := repo.files.Delete(ctx, &files.DeleteRequest{
		ID: req.ID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete")
		return nil, err
	}

	return &models.DeleteFileResponse{
		Status: removed,
	}, nil
}

func (repo *Repository) EditFile(ctx context.Context, req *models.EditFileRequest) (*models.EditFileResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "EditFile", trace.WithAttributes(
		attribute.String("file_id", req.FileID),
		attribute.String("name", req.Name),
	))
	defer span.End()

	clearName, err := utils.ClearName(req.Name)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ClearName")
		return nil, err
	}

	file, err := repo.files.Edit(ctx, &files.EditRequest{
		FileID: req.FileID,
		Name:   clearName,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Edit")
		return nil, err
	}

	return (*models.EditFileResponse)(file), nil
}

func (repo *Repository) GetFile(ctx context.Context, req *models.GetFileRequest) (*models.GetFileResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "GetFile", trace.WithAttributes(
		attribute.String("file_id", req.FileID),
	))
	defer span.End()

	file, err := repo.files.Get(ctx, &files.GetRequest{
		ID: req.FileID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFile")
		return nil, err
	}

	fileContents, err := repo.fileContent.GetMany(ctx, &file_contents.GetManyRequest{
		FileID: req.FileID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFileContents")
		return nil, err
	}

	foundAliases, err := repo.aliases.GetFileAliases(ctx, &aliases.GetFileAliasesRequest{
		FileID: req.FileID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFileAliases")
		return nil, err
	}

	return &models.GetFileResponse{
		File:     *file,
		Aliases:  foundAliases,
		Contents: fileContents,
	}, nil
}

func (repo *Repository) CreateFileContent(ctx context.Context, req *models.CreateFileContentRequest) (*models.CreateFileContentResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "CreateFileContent", trace.WithAttributes(
		attribute.String("file_id", req.FileID),
		attribute.String("version", req.Version),
		attribute.String("format_id", req.FormatID),
	))
	defer span.End()

	filesContent, err := repo.fileContent.Create(ctx, &file_contents.CreateRequest{
		FileID:   req.FileID,
		Content:  req.Content,
		Version:  req.Version,
		FormatID: req.FormatID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "CreateFileContent")
		return nil, err
	}

	return (*models.CreateFileContentResponse)(filesContent), nil
}

func (repo *Repository) GetFileContents(ctx context.Context, req *models.GetFileContentsRequest) (*models.GetFileContentsResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "GetFileContents", trace.WithAttributes(
		attribute.String("file_id", req.FileID),
	))
	defer span.End()

	filesContent, err := repo.fileContent.GetMany(ctx, &file_contents.GetManyRequest{
		FileID:  req.FileID,
		Version: req.Version,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFileContents")
		return nil, err
	}

	return (*models.GetFileContentsResponse)(&filesContent), nil
}

func (repo *Repository) EditFileContent(ctx context.Context, req *models.EditFileContentRequest) (*models.EditFileContentResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "EditFileContent", trace.WithAttributes(
		attribute.String("content_id", req.ContentID),
	))
	defer span.End()

	filesContent, err := repo.fileContent.Edit(ctx, &file_contents.EditRequest{
		FileContentID: req.ContentID,
		Content:       req.Content,
		Version:       req.Version,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "EditFileContent")
		return nil, err
	}

	go repo.callback.Send(&callback.CallbackRequest{
		FileID: filesContent.FileID,
	})
	return (*models.EditFileContentResponse)(filesContent), nil
}

func (repo *Repository) DeleteFileContent(ctx context.Context, req *models.DeleteFileContentRequest) (*models.DeleteFileContentResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "DeleteFileContent", trace.WithAttributes(
		attribute.String("content_id", req.ContentID),
	))
	defer span.End()

	removed, err := repo.fileContent.Delete(ctx, &file_contents.DeleteRequest{
		ID: req.ContentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete")
		return nil, err
	}

	return &models.DeleteFileContentResponse{
		Status: removed,
	}, nil
}

func (repo *Repository) CreateListener(ctx context.Context, req *models.CreateListenerRequest) (*models.CreateListenerResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "CreateListener", trace.WithAttributes(
		attribute.String("id", req.FileID),
	))
	defer span.End()

	listener, err := repo.listeners.Create(ctx, &listeners.CreateRequest{
		FileID:           req.FileID,
		Name:             req.Name,
		CallbackEndpoint: req.CallbackEndpoint,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Create")
		return nil, err
	}

	return (*models.CreateListenerResponse)(listener), nil
}

func (repo *Repository) GetListener(ctx context.Context, req *models.GetListenerRequest) (*models.GetListenerResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "GetListener", trace.WithAttributes(
		attribute.String("listener_id", req.ListenerID),
	))
	defer span.End()

	listener, err := repo.listeners.Get(ctx, &listeners.GetRequest{
		ID: req.ListenerID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetListener")
		return nil, err
	}

	return (*models.GetListenerResponse)(listener), nil
}

func (repo *Repository) GetFileListeners(ctx context.Context, req *models.GetFileListenersRequest) (*models.GetFileListenersResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "GetFileListeners", trace.WithAttributes(
		attribute.String("id", req.FileID),
	))
	defer span.End()

	listeners, err := repo.listeners.GetMany(ctx, &listeners.GetManyRequest{
		FileID: req.FileID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetMany")
		return nil, err
	}

	return (*models.GetFileListenersResponse)(&listeners), nil
}

func (repo *Repository) EditListener(ctx context.Context, req *models.EditListenerRequest) (*models.EditListenerResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "EditListener", trace.WithAttributes(
		attribute.String("id", req.ListenerID),
	))
	defer span.End()

	listener, err := repo.listeners.Edit(ctx, &listeners.EditRequest{
		ID:               req.ListenerID,
		Name:             req.Name,
		CallbackEndpoint: req.CallbackEndpoint,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Edit")
		return nil, err
	}

	return (*models.EditListenerResponse)(listener), nil
}

func (repo *Repository) DeleteListener(ctx context.Context, req *models.DeleteListenerRequest) (*models.DeleteListenerResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "DeleteListener", trace.WithAttributes(
		attribute.String("listener_id", req.ListenerID),
	))
	defer span.End()

	removed, err := repo.listeners.Delete(ctx, &listeners.DeleteRequest{
		ID: req.ListenerID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete")
		return nil, err
	}

	return &models.DeleteListenerResponse{
		Status: removed,
	}, nil
}

func (repo *Repository) GetContentFormats(ctx context.Context, req *models.GetContentFormatsRequest) (*models.GetContentFormatsResponse, tiny_errors.ErrorHandler) {
	repo.log.InfoContext(context.Background(), TracerName)
	ctx, span := repo.tracer.Start(context.Background(), "GetContentFormats")
	defer span.End()

	contentFormats, err := repo.contentFormats.GetMany(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetMany")
		return nil, err
	}

	return (*models.GetContentFormatsResponse)(&contentFormats), nil
}

func (repo *Repository) CreateAlias(ctx context.Context, req *models.CreateAliasRequest) (*models.CreateAliasResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "CreateAlias")
	if req == nil {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_BodyRequired)
	}
	ctx, span := repo.tracer.Start(ctx, "CreateAlias", trace.WithAttributes(
		attribute.String("key", req.Key),
		attribute.String("value", req.Key),
		attribute.String("color", req.Color),
	))
	defer span.End()

	exists, err := repo.aliases.Exists(ctx, &aliases.ExistsRequest{
		Key:   req.Key,
		Value: req.Value,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ExistsAlias")
		return nil, err
	}

	if exists {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists)
	}

	alias, err := repo.aliases.Create(ctx, &aliases.CreateRequest{
		Key:   req.Key,
		Value: req.Value,
		Color: req.Color,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "NewAlias")
		return nil, err
	}

	return (*models.CreateAliasResponse)(alias), nil
}

func (repo *Repository) GetAliases(ctx context.Context, req *models.GetAliasesRequest) (*models.GetAliasesResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "GetAliases")
	ctx, span := repo.tracer.Start(ctx, "GetAliases")
	defer span.End()

	getRequest := &aliases.GetManyRequest{}

	if req != nil {
		getRequest = &aliases.GetManyRequest{
			Key:       req.Key,
			Value:     req.Value,
			Limit:     req.Limit,
			Offset:    req.Offset,
			OrderBy:   req.OrderBy,
			OrderType: req.OrderType,
		}
	}

	aliases, err := repo.aliases.GetMany(ctx, getRequest)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetMany")
		return nil, err
	}

	return (*models.GetAliasesResponse)(&aliases), nil
}

func (repo *Repository) GetAlias(ctx context.Context, req *models.GetAliasRequests) (*models.GetAliasResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "GetAlias")
	ctx, span := repo.tracer.Start(ctx, "GetAlias")
	defer span.End()

	alias, err := repo.aliases.Get(ctx, &aliases.GetRequest{
		AliasID: req.AliasID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Get")
		return nil, err
	}

	return (*models.GetAliasResponse)(alias), nil
}

func (repo *Repository) EditAlias(ctx context.Context, req *models.EditAliasRequest) (*models.EditAliasResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "EditAlias")
	ctx, span := repo.tracer.Start(ctx, "EditAlias")
	defer span.End()

	alias, err := repo.aliases.Edit(ctx, &aliases.EditRequest{
		AliasID: req.AliasID,
		Key:     req.Key,
		Value:   req.Value,
		Color:   req.Color,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Edit")
		return nil, err
	}

	return (*models.EditAliasResponse)(alias), nil
}

func (repo *Repository) DeleteAlias(ctx context.Context, req *models.DeleteAliasRequest) (*models.DeleteAliasResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "DeleteAlias")
	ctx, span := repo.tracer.Start(ctx, "DeleteAlias")
	defer span.End()

	removed, err := repo.aliases.Delete(ctx, &aliases.DeleteRequest{
		AliasID: req.AliasID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete")
		return nil, err
	}

	return &models.DeleteAliasResponse{
		Status: removed,
	}, nil
}

func (repo *Repository) AddAliasToFile(ctx context.Context, req *models.AddAliasToFileRequest) (*models.AddAliasToFileResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "AddAliasToFile")
	ctx, span := repo.tracer.Start(ctx, "AddAliasToFile")
	defer span.End()

	exists, err := repo.aliases.ExistsInFile(ctx, &aliases.ExistsInFileRequest{
		FileID:  req.FileID,
		Aliases: req.Aliases,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "ExistsInFile")
		return nil, err
	}

	totalAdded := len(req.Aliases) - len(exists)
	if totalAdded == 0 {
		return nil, tiny_errors.New(custom_errors.ERR_CODE_Exists, tiny_errors.Message("provided aliases already exists"))
	}

	notExists := make([]string, totalAdded)
	currentIndex := 0
	for _, alias := range req.Aliases {
		if !slices.Contains(exists, alias) {
			notExists[currentIndex] = alias
			currentIndex++
		}
	}

	added, err := repo.aliases.AddToFile(ctx, &aliases.AddToFileRequest{
		FileID:  req.FileID,
		Aliases: notExists,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "AddToFile")
		return nil, err
	}

	return &models.AddAliasToFileResponse{
		Added: added,
	}, nil
}

func (repo *Repository) GetFileAliases(ctx context.Context, req *models.GetFileAliasesRequest) (*models.GetFileAliasesResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "GetFileAliases")
	ctx, span := repo.tracer.Start(ctx, "GetFileAliases")
	defer span.End()

	aliases, err := repo.aliases.GetFileAliases(ctx, &aliases.GetFileAliasesRequest{
		FileID: req.FileID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFileAliases")
		return nil, err
	}

	return (*models.GetFileAliasesResponse)(&aliases), nil
}

func (repo *Repository) RemoveFileAliases(ctx context.Context, req *models.RemoveAliasFromFileRequest) (*models.RemoveAliasFromFileResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req, "callback", "RemoveFileAliases")
	ctx, span := repo.tracer.Start(ctx, "RemoveFileAliases")
	defer span.End()

	removed, err := repo.aliases.RemoveFromFile(ctx, &aliases.RemoveFromFileRequest{
		FileID:  req.FileID,
		Aliases: req.Aliases,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "RemoveFromFile")
		return nil, err
	}

	return &models.RemoveAliasFromFileResponse{
		Removed: removed,
	}, nil
}
