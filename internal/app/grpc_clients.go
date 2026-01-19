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
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"

	"github.com/mikhail5545/product-service-client/client"
	"go.uber.org/zap"
)

type GRPCClients struct {
	VideoSvcClient *client.VideoServiceClient
	ImageSvcClient *client.ImageServiceClient
}

func setupGRPCClients(ctx context.Context, videoSvcAddr, imageSvcAddr string, logger *zap.Logger) (*GRPCClients, error) {
	videoClient, err := client.NewVideoServiceClient(client.WithTimeout(10, time.Second))
	if err != nil {
		logger.Error("failed to create Video Service gRPC client", zap.Error(err))
		return nil, err
	}
	imageClient, err := client.NewImageServiceClient(client.WithTimeout(10, time.Second))
	if err != nil {
		logger.Error("failed to create Image Service gRPC client", zap.Error(err))
		return nil, err
	}

	// Connect to Server
	pool := x509.NewCertPool()
	caPEM, err := os.ReadFile("/config/certs/ca.pem")
	if err != nil {
		logger.Error("failed to read CA certificate", zap.Error(err))
		return nil, err
	}
	pool.AppendCertsFromPEM(caPEM)
	clientCert, err := tls.LoadX509KeyPair("/config/certs/client.crt", "/config/certs/client.key")
	if err != nil {
		logger.Error("failed to load client certificate", zap.Error(err))
		return nil, err
	}
	tlsConfig := &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{clientCert},
	}
	if err := videoClient.Connect(ctx, videoSvcAddr, client.WithTLSConfig(tlsConfig)); err != nil {
		logger.Error("failed to connect to Video Service gRPC server", zap.Error(err))
		return nil, err
	}
	if err := imageClient.Connect(ctx, imageSvcAddr, client.WithTLSConfig(tlsConfig)); err != nil {
		logger.Error("failed to connect to Image Service gRPC server", zap.Error(err))
		return nil, err
	}
	return &GRPCClients{
		VideoSvcClient: videoClient,
		ImageSvcClient: imageClient,
	}, nil
}
