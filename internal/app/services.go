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
	cldservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
	"go.uber.org/zap"
)

type Services struct {
	MuxSvc *muxservice.Service
	CldSvc *cldservice.Service
}

func setupServices(repos *Repositories, apiClients *ApiClients, grpcClients *GRPCClients, logger *zap.Logger) *Services {
	return &Services{
		MuxSvc: muxservice.New(
			&muxservice.NewParams{
				Repo:         repos.Postgres.MuxRepo,
				MetadataRepo: repos.Mongo.MuxMetaRepo,
				ApiClient:    apiClients.MuxClient,
				VideoClient:  grpcClients.VideoSvcClient,
			},
			logger),
		CldSvc: cldservice.New(
			&cldservice.NewParams{
				Repo:               repos.Postgres.CldRepo,
				MetadataRepo:       repos.Mongo.CldMetaRepo,
				ApiClient:          apiClients.CldClient,
				ImageServiceClient: grpcClients.ImageSvcClient,
			}, logger),
	}
}
