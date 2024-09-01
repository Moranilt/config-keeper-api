package service

import (
	"net/http"

	"github.com/Moranilt/config-keeper/repository"
	"github.com/Moranilt/http-utils/handler"
	"github.com/Moranilt/http-utils/logger"
)

type FolderService interface {
	CreateFolder(http.ResponseWriter, *http.Request)
	GetFolder(w http.ResponseWriter, r *http.Request)
	DeleteFolder(w http.ResponseWriter, r *http.Request)
	EditFolder(w http.ResponseWriter, r *http.Request)
}

type FileService interface {
	CreateFile(w http.ResponseWriter, r *http.Request)
	DeleteFile(w http.ResponseWriter, r *http.Request)
	EditFile(w http.ResponseWriter, r *http.Request)
	GetFile(w http.ResponseWriter, r *http.Request)
}

type FileContentServices interface {
	CreateFileContent(w http.ResponseWriter, r *http.Request)
	GetFileContents(w http.ResponseWriter, r *http.Request)
	EditFileContent(w http.ResponseWriter, r *http.Request)
	DeleteFileContent(w http.ResponseWriter, r *http.Request)
}

type ListenersService interface {
	CreateListener(w http.ResponseWriter, r *http.Request)
	GetListener(w http.ResponseWriter, r *http.Request)
	GetFileListeners(w http.ResponseWriter, r *http.Request)
	EditListener(w http.ResponseWriter, r *http.Request)
	DeleteListener(w http.ResponseWriter, r *http.Request)
}

type AliasesService interface {
	CreateAlias(w http.ResponseWriter, r *http.Request)
	GetAliases(w http.ResponseWriter, r *http.Request)
	GetAlias(w http.ResponseWriter, r *http.Request)
	EditAlias(w http.ResponseWriter, r *http.Request)
	DeleteAlias(w http.ResponseWriter, r *http.Request)
	AddAliasToFile(w http.ResponseWriter, r *http.Request)
	GetFileAliases(w http.ResponseWriter, r *http.Request)
	RemoveFileAliases(w http.ResponseWriter, r *http.Request)
}

type ContentFormatsService interface {
	GetContentFormats(w http.ResponseWriter, r *http.Request)
}

type Service interface {
	FolderService
	FileService
	FileContentServices
	ListenersService
	AliasesService
	ContentFormatsService
}

type service struct {
	log  logger.Logger
	repo *repository.Repository
}

func New(log logger.Logger, repo *repository.Repository) Service {
	return &service{
		log:  log,
		repo: repo,
	}
}

func (s *service) CreateFolder(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.CreateFolder).
		WithJSON().
		Run(http.StatusCreated)
}

func (s *service) GetFolder(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetFolder).
		WithVars().
		WithQuery().
		Run(http.StatusOK)
}

func (s *service) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.DeleteFolder).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) EditFolder(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.EditFolder).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}

func (s *service) CreateFile(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.CreateFile).
		WithJSON().
		Run(http.StatusCreated)
}

func (s *service) DeleteFile(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.DeleteFile).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) EditFile(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.EditFile).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}

func (s *service) CreateFileContent(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.CreateFileContent).
		WithVars().
		WithJSON().
		Run(http.StatusCreated)
}

func (s *service) GetFile(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetFile).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) GetFileContents(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetFileContents).
		WithVars().
		WithQuery().
		Run(http.StatusOK)
}

func (s *service) EditFileContent(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.EditFileContent).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}

func (s *service) DeleteFileContent(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.DeleteFileContent).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) CreateListener(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.CreateListener).
		WithVars().
		WithJSON().
		Run(http.StatusCreated)
}

func (s *service) GetListener(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetListener).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) GetFileListeners(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetFileListeners).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) EditListener(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.EditListener).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}

func (s *service) DeleteListener(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.DeleteListener).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) GetContentFormats(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetContentFormats).
		Run(http.StatusOK)
}

func (s *service) CreateAlias(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.CreateAlias).
		WithJSON().
		Run(http.StatusCreated)
}

func (s *service) GetAliases(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetAliases).
		WithQuery().
		Run(http.StatusOK)
}

func (s *service) GetAlias(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetAlias).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) EditAlias(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.EditAlias).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}

func (s *service) DeleteAlias(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.DeleteAlias).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) AddAliasToFile(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.AddAliasToFile).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}

func (s *service) GetFileAliases(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.GetFileAliases).
		WithVars().
		Run(http.StatusOK)
}

func (s *service) RemoveFileAliases(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.RemoveFileAliases).
		WithVars().
		WithJSON().
		Run(http.StatusOK)
}
