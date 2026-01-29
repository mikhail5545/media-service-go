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
	cldmetarepo "github.com/mikhail5545/media-service-go/internal/database/mongo/cloudinary/metadata"
	muxmetarepo "github.com/mikhail5545/media-service-go/internal/database/mongo/mux/metadata"
	cldassetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/cloudinary/asset"
	muxassetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/mux/asset"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

type Repositories struct {
	Postgres *PostgresRepositories
	Mongo    *MongoRepositories
}

type PostgresRepositories struct {
	MuxRepo *muxassetrepo.Repository
	CldRepo *cldassetrepo.Repository
}

type MongoRepositories struct {
	MuxMetaRepo *muxmetarepo.Repository
	CldMetaRepo *cldmetarepo.Repository
}

func (a *App) setupRepositories() *Repositories {
	postgresRepos := setupPostgresRepositories(a.postgresDB)
	mongoRepos := setupMongoRepositories(a.mongoDB)
	return &Repositories{
		Postgres: postgresRepos,
		Mongo:    mongoRepos,
	}
}

func setupPostgresRepositories(db *gorm.DB) *PostgresRepositories {
	return &PostgresRepositories{
		MuxRepo: muxassetrepo.New(db),
		CldRepo: cldassetrepo.New(db),
	}
}

func setupMongoRepositories(db *mongo.Database) *MongoRepositories {
	return &MongoRepositories{
		MuxMetaRepo: muxmetarepo.New(db, "mux_metadata"),
		CldMetaRepo: cldmetarepo.New(db, "cloudinary_metadata"),
	}
}
