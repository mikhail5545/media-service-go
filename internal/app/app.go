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

package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	muxapi "github.com/mikhail5545/media-service-go/internal/clients/mux"
	"github.com/mikhail5545/media-service-go/internal/database"
	"github.com/mikhail5545/media-service-go/internal/database/arango"
	arangocldmetadata "github.com/mikhail5545/media-service-go/internal/database/arango/cloudinary/metadata"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/mux/asset"
	muxserver "github.com/mikhail5545/media-service-go/internal/server/mux"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
	"google.golang.org/grpc"
)

func Startup(ctx context.Context) {
	const grpcPort = 50053
	const httpPort = 8083
	grpcListenAddr := fmt.Sprintf(":%d", grpcPort)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Init postgres db connection
	DBHost := os.Getenv("POSTGRES_HOST")
	DBPort := os.Getenv("POSTGRES_PORT")
	DBUser := os.Getenv("POSTGRES_USER")
	DBPassword := os.Getenv("POSTGRES_PASSWORD")
	DBName := os.Getenv("POSTGRES_DB")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", DBHost, DBPort, DBUser, DBPassword, DBName)

	db, err := database.NewPostgresDB(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established.")

	// Init arango DB connection
	arangoDB, err := arango.NewArangoDB(ctx, []string{""})
	if err != nil {
		log.Fatalf("failed to connect to arango db: %w", err)
		os.Exit(1)
	}

	cldMetadataRepo := arangocldmetadata.New(arangoDB)
	if err := cldMetadataRepo.EnsureCollection(ctx, arangoDB); err != nil {
		log.Fatalf("Failed to ensure ArangoDB collection for cloudinary metadata: %w", err)
	}
	log.Println("ArangoDB collections initialized.")

	// Create instances of required clients
	muxClient, err := muxapi.NewMUXClient()
	if err != nil {
		log.Fatalf("Failed to create MUX client: %v", err)
	}

	// Create instances of required repositories
	muxRepo := assetrepo.New(db)

	// Create instances of required services
	muxService := muxservice.New(muxRepo, muxClient)

	// --- Start gRPC server ---
	go func() {
		lis, err := net.Listen("tcp", grpcListenAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()

		muxserver.Register(grpcServer, muxService)

		log.Printf("gRPC server listening on %s", grpcListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// --- Start HTTP server ---
	e := echo.New()

	// Setup router

	httpListenAddr := fmt.Sprintf(":%d", httpPort)
	if err := e.Start(httpListenAddr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
