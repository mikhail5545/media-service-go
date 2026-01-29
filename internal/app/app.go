/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1password/onepassword-sdk-go"
	"github.com/labstack/echo/v4"
	"github.com/mikhail5545/media-service-go/internal/app/credentials"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	Cfg         *Config
	manager     *credentials.Manager
	logger      *zap.Logger
	postgresDB  *gorm.DB
	mongoDB     *mongo.Database
	opClient    *onepassword.Client
	repos       *Repositories
	apiClients  *ApiClients
	services    *Services
	grpcClients *GRPCClients
	cleanup     func()
}

func New(ctx context.Context, cfg *Config) (*App, error) {
	logger, cleanup, err := newLogger(cfg.Log)
	if err != nil {
		return nil, err
	}

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	manager, err := credentials.New(
		ctx,
		credentials.LoadSources(),
		token,
		logger,
	)
	if err != nil {
		return nil, err
	}

	return &App{
		Cfg:     cfg,
		manager: manager,
		logger:  logger,
		cleanup: cleanup,
	}, nil
}

func (a *App) Init(ctx context.Context) error {
	if err := a.manager.ResolveAll(ctx); err != nil {
		return err
	}

	postgresDB, err := a.setupPostgresDB(ctx)
	if err != nil {
		return err
	}
	mongoDB, err := a.setupMongoDB(ctx)
	if err != nil {
		return err
	}

	repos := a.setupRepositories()

	apiClients, err := a.setupApiClients()
	if err != nil {
		return err
	}

	grpcClients, err := a.setupGRPCClients(ctx)
	if err != nil {
		return err
	}

	services := a.setupServices(repos, apiClients, grpcClients, a.logger)

	a.postgresDB = postgresDB
	a.mongoDB = mongoDB
	a.repos = repos
	a.apiClients = apiClients
	a.services = services

	return nil
}

func (a *App) Run(ctx context.Context) error {
	grpcServer, listener, err := a.prepareGRPCServer()
	if err != nil {
		return err
	}

	grpcErrChan := make(chan error, 1)
	go runGRPCServer(grpcErrChan, grpcServer, listener, a.logger)

	e := echo.New()
	integrateWithEcho(e, a.logger)
	setupRouters(e, a.services)

	httpErrChan := make(chan error, 1)
	go runHTTPServer(e, a.Cfg.HTTP.Port, a.logger, httpErrChan)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-quit:
		a.logger.Info("received shutdown signal", zap.String("signal", sig.String()))
	case <-ctx.Done():
		a.logger.Info("received shutdown signal via context cancellation")
	case err := <-grpcErrChan:
		a.logger.Error("gRPC server stopped unexpectedly with an error", zap.Error(err))
	case err := <-httpErrChan:
		a.logger.Error("gRPC server stopped unexpectedly with an error", zap.Error(err))
	}

	a.gracefulShutdown(e, grpcServer, a.logger)
	return nil
}

func (a *App) gracefulShutdown(e *echo.Echo, grpcServer *grpc.Server, logger *zap.Logger) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Cfg.GracefulShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	shutdownHTTPServer(shutdownCtx, e, logger)

	done := make(chan struct{})
	go shutdownGRPCServer(grpcServer, done)

	select {
	case <-done:
		logger.Info("gRPC server shutdown complete")
	case <-shutdownCtx.Done():
		logger.Warn("gRPC graceful shutdown timed out, forcing stop")
		grpcServer.Stop()
	}
}

func (a *App) Close() error {
	if a.cleanup != nil {
		a.cleanup()
	}
	if a.grpcClients != nil {
		if err := a.grpcClients.VideoSvcClient.Close(); err != nil {
			return err
		}
		if err := a.grpcClients.ImageSvcClient.Close(); err != nil {
			return err
		}
	}
	return nil
}
