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

package mux

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	mux "github.com/muxinc/mux-go/v6"
)

type APIClient interface {
	CreateDirectUploadURL(ctx context.Context, meta *mux.AssetMetadata, policies ...mux.PlaybackPolicy) (*mux.UploadResponse, error)
	DeleteAsset(ctx context.Context, assetID string) error
}

type Client struct {
	client *mux.APIClient
	cfg    config
}

var _ APIClient = (*Client)(nil)

func New(apiKey, secretKey string, opt ...Option) (*Client, error) {
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("api key or secret key is empty")
	}

	client := mux.NewAPIClient(mux.NewConfiguration(
		mux.WithBasicAuth(apiKey, secretKey),
	))

	cfg := &config{}
	for _, o := range opt {
		if err := o(cfg); err != nil {
			return nil, fmt.Errorf("error applying option: %w", err)
		}
	}

	return &Client{
		client: client,
		cfg:    *cfg,
	}, nil
}

func (c *Client) CreateDirectUploadURL(ctx context.Context, meta *mux.AssetMetadata, policies ...mux.PlaybackPolicy) (*mux.UploadResponse, error) {
	assetReq := mux.CreateAssetRequest{
		PlaybackPolicy: policies,
		VideoQuality:   "basic",
	}
	if meta != nil {
		assetReq.Meta = *meta
	}

	if c.cfg.corsOrigin == "" {
		c.cfg.corsOrigin = "*"
	}
	uploadRequest := mux.CreateUploadRequest{
		NewAssetSettings: assetReq,
		CorsOrigin:       c.cfg.corsOrigin,
		Test:             c.cfg.test,
	}

	resp, err := c.client.DirectUploadsApi.CreateDirectUpload(uploadRequest, mux.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to create direct upload url: %w", err)
	}
	return &resp, nil
}

func (c *Client) DeleteAsset(ctx context.Context, assetID string) error {
	if err := c.client.AssetsApi.DeleteAsset(assetID, mux.WithContext(ctx)); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}
	return nil
}

type GeneratePlaybackTokenOptions struct {
	UserID     uuid.UUID
	PlaybackID string
	Expiration int64      // in seconds
	UserAgent  *string    // optional
	SessionID  *uuid.UUID // optional
}

func populateCustomClaims(opts GeneratePlaybackTokenOptions) map[string]any {
	custom := make(map[string]any)
	if opts.UserID != uuid.Nil {
		custom["user_id"] = opts.UserID.String()
	}
	if opts.UserAgent != nil {
		custom["user_agent"] = *opts.UserAgent
	}
	if opts.SessionID != nil {
		custom["session_id"] = opts.SessionID.String()
	}
	return custom
}

func (c *Client) GeneratePlaybackJWTToken(opts GeneratePlaybackTokenOptions) (string, error) {
	if len(c.cfg.signingKeyPrivateKey) == 0 || c.cfg.signingKeyID == "" {
		return "", fmt.Errorf("signing key is not configured")
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(c.cfg.signingKeyPrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse signing key: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": opts.PlaybackID,
		"aud": "v",
		"exp": opts.Expiration,
		"kid": c.cfg.signingKeyID,
	})

	custom := populateCustomClaims(opts)
	if len(custom) > 0 {
		token.Claims.(jwt.MapClaims)["custom"] = custom
	}

	singedToken, err := token.SignedString(signKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return singedToken, nil
}
