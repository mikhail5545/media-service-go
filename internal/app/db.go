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

	"github.com/mikhail5545/media-service-go/internal/database/postgres"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func setupPostgresDB(ctx context.Context, sp SecretProvider, logger *zap.Logger, dbCfg PostgresDBConfig) (*gorm.DB, error) {
	host, err := getSecret(ctx, sp, dbCfg.HostRef)
	if err != nil {
		return nil, err
	}
	port, err := getSecret(ctx, sp, dbCfg.PortRef)
	if err != nil {
		return nil, err
	}
	user, err := getSecret(ctx, sp, dbCfg.UserRef)
	if err != nil {
		return nil, err
	}
	password, err := getSecret(ctx, sp, dbCfg.PasswordRef)
	if err != nil {
		return nil, err
	}
	dbName, err := getSecret(ctx, sp, dbCfg.DBNameRef)
	if err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	logger.Info("database DSN prepared", zap.String("host", host), zap.String("port", port), zap.String("user", user))
	db, err := postgres.NewPostgresDB(ctx, dsn)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("database connection established.")
	return db, nil
}

func setupMongoClient(ctx context.Context, sp SecretProvider, logger *zap.Logger, mongoCfg MongoDBConfig) (*mongo.Client, error) {
	uri, err := getSecret(ctx, sp, mongoCfg.ConnectionStringRef)
	if err != nil {
		return nil, err
	}
	logger.Info("MongoDB URI string resolved")
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, err
	}
	return client, nil
}

func basicDBConfig() PostgresDBConfig {
	return PostgresDBConfig{
		HostRef:     "op://Development/Postgres/server",
		PortRef:     "op://Development/Postgres/port",
		UserRef:     "op://Development/Postgres/username",
		PasswordRef: "op://Development/Postgres/password",
		DBNameRef:   "op://Development/Postgres/database",
	}
}
