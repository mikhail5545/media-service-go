// github.com/mikhail5545/media-service-go
// microservice for vitianmove project family
// Copyright (C) 2025  Mikhail Kulik

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package arango

import (
	"context"
	"fmt"
	"os"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
)

func NewArangoDB(ctx context.Context, e []string) (arangodb.Database, error) {
	// Initialize arangoDB client
	endpoint := connection.NewRoundRobinEndpoints(e)
	conn := connection.NewHttp2Connection(connection.DefaultHTTP2ConfigurationWrapper(endpoint, false))
	auth := connection.NewBasicAuth("root", "password")
	if err := conn.SetAuthentication(auth); err != nil {
		return nil, fmt.Errorf("failed to set up auth for arango db connection: %w", err)
	}
	arangoClient := arangodb.NewClient(conn)

	dbName := "media_service"
	exists, err := arangoClient.DatabaseExists(ctx, dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to check for database existance: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("database '%s' does not exist: %w", dbName, err)
	}

	db, err := arangoClient.GetDatabase(ctx, dbName, &arangodb.GetDatabaseOptions{SkipExistCheck: false})
	if err != nil {
		return nil, fmt.Errorf("failed to get arango database '%s': %w", dbName, err)
	}

	return db, nil
}

func CreateArangoDB(ctx context.Context, name string, c arangodb.Client) (arangodb.Database, error) {
	dbName := "media_service"
	db, err := c.CreateDatabase(ctx, dbName, &arangodb.CreateDatabaseOptions{
		Users: []arangodb.CreateDatabaseUserOptions{
			{
				UserName: os.Getenv("ARANGO_DB_USERNAME"),
				Password: os.Getenv("ARANGO_DB_PASSWORD"),
			},
		},
		Options: arangodb.CreateDatabaseDefaultOptions{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new arango db with '%s' name: %w", dbName, err)
	}
	return db, nil
}
