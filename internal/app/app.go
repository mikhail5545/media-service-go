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
	"strings"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	cldapi "github.com/mikhail5545/media-service-go/internal/clients/cloudinary"
	muxapi "github.com/mikhail5545/media-service-go/internal/clients/mux"
	"github.com/mikhail5545/media-service-go/internal/database"
	"github.com/mikhail5545/media-service-go/internal/database/arango"
	arangocldmetadata "github.com/mikhail5545/media-service-go/internal/database/arango/cloudinary/metadata"
	arangomuxmetadata "github.com/mikhail5545/media-service-go/internal/database/arango/mux/metadata"
	cldassetrepo "github.com/mikhail5545/media-service-go/internal/database/cloudinary/asset"
	muxassetrepo "github.com/mikhail5545/media-service-go/internal/database/mux/asset"
	muxdetailrepo "github.com/mikhail5545/media-service-go/internal/database/mux/detail"
	"github.com/mikhail5545/media-service-go/internal/routers"
	cldserver "github.com/mikhail5545/media-service-go/internal/server/cloudinary"
	muxserver "github.com/mikhail5545/media-service-go/internal/server/mux"
	cldservice "github.com/mikhail5545/media-service-go/internal/services_outdated/cloudinary"
	muxservice "github.com/mikhail5545/media-service-go/internal/services_outdated/mux"
	imageclient "github.com/mikhail5545/product-service-go/pkg/client/image"
	videoclient "github.com/mikhail5545/product-service-go/pkg/client/video"
	"google.golang.org/grpc"
)

func Startup(ctx context.Context) {
	const grpcPort = 50053
	const httpPort = 8083
	grpcListenAddr := fmt.Sprintf(":%d", grpcPort)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		os.Exit(1)
	}

	// Load .env variables
	DBHost := os.Getenv("POSTGRES_HOST")
	DBPort := os.Getenv("POSTGRES_PORT")
	DBUser := os.Getenv("POSTGRES_USER")
	DBPassword := os.Getenv("POSTGRES_PASSWORD")
	DBName := os.Getenv("POSTGRES_DB")

	videoClientAddr := os.Getenv("VIDEO_SERVICE_ADDR")
	imageClientAddr := os.Getenv("IMAGE_SERVICE_ADDR")
	arangoDBEndpoints := strings.Split(os.Getenv("ARANGO_DB_ENDPOINTS"), "|")

	// Init postgres db connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", DBHost, DBPort, DBUser, DBPassword, DBName)

	db, err := database.NewPostgresDB(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %s", err.Error())
		os.Exit(1)
	}

	log.Println("Database connection established.")

	// Init arango DB connection
	arangoDB, err := arango.NewArangoDB(ctx, arangoDBEndpoints)
	if err != nil {
		log.Fatalf("failed to connect to arango db: %s", err.Error())
		os.Exit(1)
	}

	cldMetadataRepo := arangocldmetadata.New(arangoDB)
	if err := cldMetadataRepo.EnsureCollection(ctx, arangoDB); err != nil {
		log.Fatalf("Failed to ensure ArangoDB collection for cloudinary metadata: %s", err.Error())
	}
	muxMetadataRepo := arangomuxmetadata.New(arangoDB)
	if err := cldMetadataRepo.EnsureCollection(ctx, arangoDB); err != nil {
		log.Fatalf("Failed to ensure ArangoDB collection for mux metadata: %s", err.Error())
	}
	log.Println("ArangoDB collections initialized.")

	// Create instances of required clients
	muxClient, err := muxapi.NewMUXClient()
	if err != nil {
		log.Fatalf("Failed to create MUX client: %s", err.Error())
		os.Exit(1)
	}
	cldClient, err := cldapi.NewCloudinaryClient()
	if err != nil {
		log.Fatalf("Failed to create Cloudinary client: %s", err.Error())
		os.Exit(1)
	}

	// Connect to required gRPC clients
	videoSvcClient, err := videoclient.New(ctx, videoClientAddr)
	imageSvcClient, err := imageclient.New(ctx, imageClientAddr)

	// Create instances of required repositories
	muxRepo := muxassetrepo.New(db)
	muxDetailRepo := muxdetailrepo.New(db)
	cldRepo := cldassetrepo.New(db)

	// Create instances of required services
	muxService := muxservice.New(muxRepo, muxMetadataRepo, muxDetailRepo, muxClient, videoSvcClient)
	cldService := cldservice.New(cldClient, cldRepo, cldMetadataRepo, imageSvcClient)

	// --- Start gRPC server ---
	go func() {
		lis, err := net.Listen("tcp", grpcListenAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %s", err.Error())
			os.Exit(1)
		}
		grpcServer := grpc.NewServer()

		muxserver.Register(grpcServer, muxService)
		cldserver.Register(grpcServer, cldService)

		log.Printf("gRPC server listening on %s", grpcListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %s", err.Error())
			os.Exit(1)
		}
	}()

	// --- Start HTTP server ---
	e := echo.New()

	// Setup router
	routers.SetupRouter(e, muxService, cldService)

	httpListenAddr := fmt.Sprintf(":%d", httpPort)
	if err := e.Start(httpListenAddr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
		os.Exit(1)
	}
}
