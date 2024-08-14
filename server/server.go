package server

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/Moranilt/config-keeper/config"
	"github.com/Moranilt/config-keeper/custom_errors"
	"github.com/Moranilt/config-keeper/endpoints"
	"github.com/Moranilt/config-keeper/middleware"
	"github.com/Moranilt/config-keeper/pkg/callback"
	"github.com/Moranilt/config-keeper/pkg/file_contents"
	"github.com/Moranilt/config-keeper/pkg/files"
	"github.com/Moranilt/config-keeper/pkg/folders"
	"github.com/Moranilt/config-keeper/pkg/listeners"
	"github.com/Moranilt/config-keeper/repository"
	"github.com/Moranilt/config-keeper/service"
	"github.com/Moranilt/config-keeper/tracer"
	"github.com/Moranilt/config-keeper/transport"
	"github.com/Moranilt/http-utils/client"
	"github.com/Moranilt/http-utils/clients/database"
	"github.com/Moranilt/http-utils/logger"
	"github.com/Moranilt/http-utils/tiny_errors"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"golang.org/x/sync/errgroup"
)

const (
	DB_DRIVER_NAME = "postgres"
)

func Run(ctx context.Context) {
	log := logger.New(os.Stdout, logger.TYPE_DEFAULT)
	logger.SetDefault(log)
	tiny_errors.Init(custom_errors.ERRORS)

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.New(ctx, DB_DRIVER_NAME, cfg.DB, cfg.Production)
	if err != nil {
		log.Fatalf("db connection: %v", err)
	}
	defer db.Close()

	// Tracer
	tp, err := tracer.NewProvider(cfg.Tracer.URL, cfg.Tracer.Name)
	if err != nil {
		log.Fatalf("tracer: %v", err)
	}

	// Migrations
	err = RunMigrations(log, db.DB.DB, cfg.DB.DBName)
	if err != nil {
		log.Fatalf("migration: %v", err)
	}

	foldersClient := folders.New(db)
	filesClient := files.New(db)
	fileContentClient := file_contents.New(db)
	listenersClient := listeners.New(db)

	callbackChannel := callback.NewChannel()

	repo := repository.New(db, callbackChannel, foldersClient, filesClient, fileContentClient, listenersClient, log)
	svc := service.New(log, repo)
	mw := middleware.New(log)
	ep := endpoints.MakeEndpoints(svc, mw)
	health := endpoints.MakeHealth(db)
	ep = append(ep, health)
	server := transport.New(fmt.Sprintf(":%s", cfg.Port), ep, mw)

	httpClient := client.New()
	client.SetTimeout(60 * time.Second)
	requestsController := callback.NewRequestsController(log, httpClient)
	callbackService := callback.New(log, callbackChannel, filesClient, listenersClient, fileContentClient, requestsController)
	go callbackService.Run(ctx)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-gCtx.Done()
		return tp.Shutdown(context.Background())
	})

	g.Go(func() error {
		<-gCtx.Done()
		return server.Shutdown(context.Background())
	})

	g.Go(func() error {
		return server.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Infof("exit with: %s", err)
	}
}

func RunMigrations(log logger.Logger, db *sql.DB, databaseName string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", databaseName, driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	version, _, err := m.Version()
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("migration: version %d", version))
	return nil
}
