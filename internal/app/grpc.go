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
	"fmt"
	"net"
	"strconv"

	"github.com/mikhail5545/media-service-go/internal/grpc/cloudinary"
	"github.com/mikhail5545/media-service-go/internal/grpc/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func registerGRPCServices(server *grpc.Server, services *Services, logger *zap.Logger) {
	mux.Register(server, services.MuxSvc, logger)
	cloudinary.Register(server, services.CldSvc, logger)
}

func (a *App) prepareGRPCServer() (*grpc.Server, net.Listener, error) {
	grpcListenAddr := ":" + strconv.FormatInt(a.Cfg.GRPC.Port, 10)
	list, err := net.Listen("tcp", grpcListenAddr)
	if err != nil {
		a.logger.Error("failed to listen on gRPC address", zap.String("address", grpcListenAddr), zap.Error(err))
		return nil, nil, fmt.Errorf("failed to listen on gRPC address %s: %w", grpcListenAddr, err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(a.manager.Credentials.GRPCServer.Credentials))
	registerGRPCServices(grpcServer, a.services, a.logger)
	return grpcServer, list, nil
}

func runGRPCServer(errChan chan<- error, grpcServer *grpc.Server, listener net.Listener, logger *zap.Logger) {
	logger.Info("starting gRPC server", zap.String("address", listener.Addr().String()))
	if err := grpcServer.Serve(listener); err != nil {
		errChan <- err
	} else {
		errChan <- nil
	}
}

func shutdownGRPCServer(grpcServer *grpc.Server, done chan<- struct{}) {
	grpcServer.GracefulStop()
	close(done)
}
