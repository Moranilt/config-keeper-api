package service

import (
	"net/http"

	"github.com/Moranilt/config-keeper/repository"
	"github.com/Moranilt/http-utils/handler"
	"github.com/Moranilt/http-utils/logger"
)

type Service interface {
	CreateUser(http.ResponseWriter, *http.Request)
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

func (s *service) CreateUser(w http.ResponseWriter, r *http.Request) {
	handler.New(w, r, s.log, s.repo.CreateUser).
		WithJSON().
		Run(http.StatusOK)
}
