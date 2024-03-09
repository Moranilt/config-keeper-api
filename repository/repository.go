package repository

import (
	"context"
	"time"

	"github.com/Moranilt/config-keeper/models"
	"github.com/Moranilt/http-utils/logger"
	"github.com/Moranilt/http-utils/tiny_errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	QUERY_InsertUser = "INSERT INTO test (firstname, lastname, patronymic) VALUES ($1, $2, $3) RETURNING id"
)

const (
	REDIS_TTL = 30 * time.Second
)

const TracerName string = "repository"

type Repository struct {
	db  *mongo.Database
	log logger.Logger
}

func New(db *mongo.Database, logger logger.Logger) *Repository {
	return &Repository{
		db:  db,
		log: logger,
	}
}

func (repo *Repository) CreateUser(ctx context.Context, req *models.TestRequest) (*models.TestResponse, tiny_errors.ErrorHandler) {
	repo.log.WithRequestId(ctx).InfoContext(ctx, TracerName, "data", req)
	_, span := otel.Tracer(TracerName).Start(ctx, "Test", trace.WithAttributes(
		attribute.String("Firstname", req.Firstname),
		attribute.String("Lastname", req.Lastname),
		attribute.String("Patronymic", *req.Patronymic),
	))
	defer span.End()

	return &models.TestResponse{
		ID: "1",
	}, nil
}
