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

	cldapiclient "github.com/mikhail5545/media-service-go/internal/apiclients/cloudinary"
	muxapiclient "github.com/mikhail5545/media-service-go/internal/apiclients/mux"
	"go.uber.org/zap"
)

type ApiClients struct {
	MuxClient *muxapiclient.Client
	CldClient *cldapiclient.Client
}

func setupApiClients(ctx context.Context, cfg *Config, sp SecretProvider, logger *zap.Logger) (*ApiClients, error) {
	muxClient, err := setupMuxApi(ctx, cfg, sp)
	if err != nil {
		logger.Error("failed to setup Mux API client", zap.Error(err))
		return nil, err
	}
	cldClient, err := setupCloudinaryApi(ctx, cfg, sp)
	if err != nil {
		logger.Error("failed to setup Cloudinary API client", zap.Error(err))
		return nil, err
	}
	return &ApiClients{
		MuxClient: muxClient,
		CldClient: cldClient,
	}, nil
}

func setupMuxApi(ctx context.Context, cfg *Config, sp SecretProvider) (*muxapiclient.Client, error) {
	muxApiKey, err := getSecret(ctx, sp, cfg.Mux.APIKeyRef)
	if err != nil {
		return nil, err
	}
	muxSecretKey, err := getSecret(ctx, sp, cfg.Mux.SecretKeyRef)
	if err != nil {
		return nil, err
	}
	muxSigningKeyID, err := getSecret(ctx, sp, cfg.Mux.SigningKeyIDRef)
	if err != nil {
		return nil, err
	}
	muxSingingKeyPrivate, err := getSecret(ctx, sp, cfg.Mux.SigningKeyPrivateRef)
	if err != nil {
		return nil, err
	}
	muxPlaybackRestrictionID, err := getSecret(ctx, sp, cfg.Mux.PlaybackRestrictionIDRef)
	if err != nil {
		return nil, err
	}
	muxClient, err := muxapiclient.New(
		muxApiKey,
		muxSecretKey,
		muxapiclient.WithSigningKey(muxSigningKeyID, muxSingingKeyPrivate),
		muxapiclient.WithCORSOrigin(cfg.Mux.CORSOrigin),
		muxapiclient.WithTestMode(cfg.Mux.TestMode),
		muxapiclient.WithPlaybackRestrictionID(muxPlaybackRestrictionID),
	)
	return muxClient, err
}

func setupCloudinaryApi(ctx context.Context, cfg *Config, sp SecretProvider) (*cldapiclient.Client, error) {
	cloudName, err := getSecret(ctx, sp, cfg.Cloudinary.CloudNameRef)
	if err != nil {
		return nil, err
	}
	apiKey, err := getSecret(ctx, sp, cfg.Cloudinary.APIKeyRef)
	if err != nil {
		return nil, err
	}
	apiSecret, err := getSecret(ctx, sp, cfg.Cloudinary.APISecretRef)
	if err != nil {
		return nil, err
	}
	cldClient, err := cldapiclient.New(
		cloudName,
		apiKey,
		apiSecret,
	)
	return cldClient, err
}
