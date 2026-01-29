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
	"time"

	"github.com/mikhail5545/product-service-client/client"
	"go.uber.org/zap"
)

type GRPCClients struct {
	VideoSvcClient *client.VideoServiceClient
	ImageSvcClient *client.ImageServiceClient
}

func (a *App) setupGRPCClients(ctx context.Context) (*GRPCClients, error) {
	videoClient, err := client.NewVideoServiceClient(client.WithTimeout(10, time.Second))
	if err != nil {
		a.logger.Error("failed to create Video Service gRPC client", zap.Error(err))
		return nil, err
	}
	imageClient, err := client.NewImageServiceClient(client.WithTimeout(10, time.Second))
	if err != nil {
		a.logger.Error("failed to create Image Service gRPC client", zap.Error(err))
		return nil, err
	}

	if err := videoClient.Connect(ctx,
		a.manager.Credentials.GRPCClient.Address,
		client.WithTransportCredentials(a.manager.Credentials.GRPCClient.Credentials),
	); err != nil {
		a.logger.Error("failed to connect to Video Service gRPC server", zap.Error(err))
		return nil, err
	}
	if err := imageClient.Connect(ctx,
		a.manager.Credentials.GRPCClient.Address,
		client.WithTransportCredentials(a.manager.Credentials.GRPCClient.Credentials),
	); err != nil {
		a.logger.Error("failed to connect to Image Service gRPC server", zap.Error(err))
		return nil, err
	}
	return &GRPCClients{
		VideoSvcClient: videoClient,
		ImageSvcClient: imageClient,
	}, nil
}
