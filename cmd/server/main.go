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

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/mikhail5545/media-service-go/internal/clients/mux"
	"github.com/mikhail5545/media-service-go/internal/database"
	"github.com/mikhail5545/media-service-go/internal/routers"
	"github.com/mikhail5545/media-service-go/internal/services"
	"google.golang.org/grpc"
)

func main() {
	const grpcPort = 50053
	const httpPort = 8083
	grpcListenAddr := fmt.Sprintf(":%d", grpcPort)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	// Create instances of required clients
	muxClient, err := mux.NewMUXClient()
	if err != nil {
		log.Fatalf("Failed to create MUX client: %v", err)
	}

	// Create instances of required repositories
	muxRepo := database.NewMUXRepository(db)

	// Create instances of required services
	muxService := services.NewMuxService(muxRepo, muxClient)

	// --- Start gRPC server ---
	go func() {
		lis, err := net.Listen("tcp", grpcListenAddr)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		grpcServer := startGRPCServer(muxService, &lis)
		log.Printf("gRPC server listening on %s", grpcListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// --- Start HTTP server ---
	e := echo.New()

	// Setup router
	routers.SetupRouter(e, muxService)

	httpListenAddr := fmt.Sprintf(":%d", httpPort)
	if err := e.Start(httpListenAddr); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func startGRPCServer(muxService *services.MuxService, lis *net.Listener) *grpc.Server {
	grpcServer := grpc.NewServer()

	return grpcServer
}
