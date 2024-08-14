package repository

import (
	"context"

	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/models"
	"github.com/Moranilt/config-keeper/pkg/callback"
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
	db          *database.Client
	log         logger.Logger
	tracer      trace.Tracer
	folders     folders.Client
	files       files.Client
	fileContent file_contents.Client
	listeners   listeners.Client
	callback    callback.CallbackChannel
}

func New(
	db *database.Client,
	callback callback.CallbackChannel,
	folders folders.Client,
	files files.Client,
	fileContent file_contents.Client,
	listeners listeners.Client,
	logger logger.Logger,
) *Repository {
	return &Repository{
		db:          db,
		log:         logger,
		tracer:      otel.Tracer(TracerName),
		folders:     folders,
		files:       files,
		fileContent: fileContent,
		listeners:   listeners,
		callback:    callback,
	}
}

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

	folder, err := repo.folders.New(ctx, &folders.NewRequest{
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
		attribute.String("id", req.ID),
	))
	defer span.End()

	var (
		folderWithPath *folders.FolderWithPath
		err            tiny_errors.ErrorHandler
		parentID       *string
	)

	if req.ID == "root" {
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
		parentID = &req.ID
		folderWithPath, err = repo.folders.Get(ctx, &folders.GetRequest{
			ID: req.ID,
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

	files, err := repo.files.GetFilesInFolder(ctx, &files.GetFilesInFolderRequest{
		FolderID: parentID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFilesInFolder")
		return nil, err
	}

	return &models.GetFolderResponse{
		ID:        folderWithPath.ID,
		Name:      folderWithPath.Name,
		ParentID:  folderWithPath.ParentID,
		CreatedAt: folderWithPath.CreatedAt,
		UpdatedAt: folderWithPath.UpdatedAt,
		Path:      folderWithPath.Path,
		Folders:   folders,
		Files:     files,
	}, nil
}

func (repo *Repository) DeleteFolder(ctx context.Context, req *models.DeleteFolderRequest) (*models.DeleteFolderResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "DeleteFolder", trace.WithAttributes(
		attribute.String("id", req.ID),
	))
	defer span.End()

	removed, err := repo.folders.Delete(ctx, &folders.DeleteRequest{
		ID: req.ID,
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
		attribute.String("id", req.ID),
		attribute.String("name", req.Name),
	))
	defer span.End()

	requiredFields := []utils.RequiredField{
		{
			Name:  "id",
			Value: req.ID,
		},
		{
			Name:  "name",
			Value: req.Name,
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
		ID:   req.ID,
		Name: req.Name,
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
		attribute.String("id", req.ID),
		attribute.String("name", req.Name),
	))
	defer span.End()

	file, err := repo.files.Edit(ctx, &files.EditRequest{
		FileID: req.ID,
		Name:   req.Name,
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
		attribute.String("id", req.ID),
	))
	defer span.End()

	file, err := repo.files.Get(ctx, &files.GetRequest{
		ID: req.ID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFile")
		return nil, err
	}

	fileContents, err := repo.fileContent.GetMany(ctx, &file_contents.GetManyRequest{
		FileID: req.ID,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "GetFileContents")
		return nil, err
	}
	return &models.GetFileResponse{
		File:     *file,
		Contents: fileContents,
	}, nil
}

func (repo *Repository) CreateFileContent(ctx context.Context, req *models.CreateFileContentRequest) (*models.CreateFileContentResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	ctx, span := repo.tracer.Start(ctx, "CreateFileContent", trace.WithAttributes(
		attribute.String("file_id", req.FileID),
		attribute.String("version", req.Version),
	))
	defer span.End()

	filesContent, err := repo.fileContent.Create(ctx, &file_contents.CreateRequest{
		FileID:  req.FileID,
		Content: req.Content,
		Version: req.Version,
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
		attribute.String("file_id", req.ID),
	))
	defer span.End()

	filesContent, err := repo.fileContent.Edit(ctx, &file_contents.EditRequest{
		FileContentID: req.ID,
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
		attribute.String("id", req.ID),
	))
	defer span.End()

	removed, err := repo.fileContent.Delete(ctx, &file_contents.DeleteRequest{
		ID: req.ID,
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
		attribute.String("id", req.ID),
	))
	defer span.End()

	listener, err := repo.listeners.Get(ctx, &listeners.GetRequest{
		ID: req.ID,
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
		attribute.String("id", req.ID),
	))
	defer span.End()

	removed, err := repo.listeners.Delete(ctx, &listeners.DeleteRequest{
		ID: req.ID,
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
