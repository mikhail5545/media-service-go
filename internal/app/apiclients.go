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
	cldapiclient "github.com/mikhail5545/media-service-go/internal/apiclients/cloudinary"
	muxapiclient "github.com/mikhail5545/media-service-go/internal/apiclients/mux"
	"go.uber.org/zap"
)

type ApiClients struct {
	MuxClient *muxapiclient.Client
	CldClient *cldapiclient.Client
}

func (a *App) setupApiClients() (*ApiClients, error) {
	muxClient, err := a.setupMuxApi()
	if err != nil {
		a.logger.Error("failed to setup Mux API client", zap.Error(err))
		return nil, err
	}
	cldClient, err := a.setupCloudinaryApi()
	if err != nil {
		a.logger.Error("failed to setup Cloudinary API client", zap.Error(err))
		return nil, err
	}
	return &ApiClients{
		MuxClient: muxClient,
		CldClient: cldClient,
	}, nil
}

func (a *App) setupMuxApi() (*muxapiclient.Client, error) {
	muxClient, err := muxapiclient.New(
		a.manager.Credentials.MuxAPI.APIToken,
		a.manager.Credentials.MuxAPI.SecretKey,
		muxapiclient.WithSigningKey(a.manager.Credentials.MuxAPI.SigningKeyID, a.manager.Credentials.MuxAPI.SigningKeyPrivate),
		muxapiclient.WithCORSOrigin(a.Cfg.Mux.CORSOrigin),
		muxapiclient.WithTestMode(a.Cfg.Mux.TestMode),
		muxapiclient.WithPlaybackRestrictionID(a.manager.Credentials.MuxAPI.PlaybackRestrictionID),
	)
	return muxClient, err
}

func (a *App) setupCloudinaryApi() (*cldapiclient.Client, error) {
	cldClient, err := cldapiclient.New(
		a.manager.Credentials.CloudinaryAPI.CloudName,
		a.manager.Credentials.CloudinaryAPI.APIKey,
		a.manager.Credentials.CloudinaryAPI.APISecret,
	)
	return cldClient, err
}
