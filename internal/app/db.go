/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package app

import (
	"context"
	"fmt"

	mongodb "github.com/mikhail5545/media-service-go/internal/database/mongo"
	"github.com/mikhail5545/media-service-go/internal/database/postgres"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (a *App) setupPostgresDB(ctx context.Context) (*gorm.DB, error) {
	pgCfg := a.manager.Credentials.PostgresDB
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgCfg.Host, pgCfg.Port, pgCfg.User, pgCfg.User, pgCfg.Password)
	a.logger.Info("database DSN prepared", zap.String("dsn", dsn))

	db, err := postgres.NewPostgresDB(ctx, dsn)
	if err != nil {
		a.logger.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	a.logger.Info("database connection established.")
	return db, nil
}

func (a *App) setupMongoDB(ctx context.Context) (*mongo.Database, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(a.manager.Credentials.MongoDB.ConnectionString).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)
	if err != nil {
		a.logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, err
	}

	db, err := mongodb.NewMongoDB(ctx, client, a.manager.Credentials.MongoDB.DBName)
	if err != nil {
		a.logger.Error("Failed to ping MongoDB", zap.Error(err))
		return nil, err
	}
	return db, nil
}
