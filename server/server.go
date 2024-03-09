package server

import (
	"context"
	"fmt"
	"os"

	"github.com/Moranilt/config-keeper/config"
	"github.com/Moranilt/config-keeper/config/database"
	"github.com/Moranilt/config-keeper/endpoints"
	"github.com/Moranilt/config-keeper/middleware"
	"github.com/Moranilt/config-keeper/repository"
	"github.com/Moranilt/config-keeper/service"
	"github.com/Moranilt/config-keeper/tracer"
	"github.com/Moranilt/config-keeper/transport"
	"github.com/Moranilt/http-utils/clients/vault"
	"github.com/Moranilt/http-utils/logger"
	_ "github.com/golang-migrate/migrate/source/file"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context) {
	log := logger.New(os.Stdout, logger.TYPE_JSON)
	logger.SetDefault(log)

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Vault
	err = vault.Init(&vault.Config{
		MountPath: cfg.Vault.MountPath,
		Token:     cfg.Vault.Token,
		Host:      cfg.Vault.Host,
	})
	if err != nil {
		log.Fatalf("vault: %v", err)
	}

	// Database
	dbCreds, err := vault.GetCreds[database.DBCreds](ctx, cfg.Vault.DbCredsPath)
	if err != nil {
		log.Fatalf("get db creds from vault: %v", err)
	}

	dbCLient, db, err := database.New(ctx, dbCreds)
	if err != nil {
		log.Fatalf("db connection: %v", err)
	}

	// Tracer
	tp, err := tracer.NewProvider(cfg.Tracer.URL, cfg.Tracer.Name)
	if err != nil {
		log.Fatalf("tracer: %v", err)
	}

	repo := repository.New(db, log)
	svc := service.New(log, repo)
	mw := middleware.New(log)
	ep := endpoints.MakeEndpoints(svc, mw)
	health := endpoints.MakeHealth(database.MakeChecker(dbCLient))
	ep = append(ep, health)
	server := transport.New(fmt.Sprintf(":%s", cfg.Port), ep, mw)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-gCtx.Done()
		return tp.Shutdown(context.Background())
	})

	g.Go(func() error {
		<-gCtx.Done()
		return dbCLient.Disconnect(context.Background())
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
