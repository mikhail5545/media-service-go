// github.com/mikhail5545/product-service-go
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

package mux

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	mux "github.com/muxinc/mux-go/v6"
)

type MUX interface {
	// CreateUploadURL creates url for direct upload to the mux using mux API.
	// It also sets metadata for the created asset using mux assset's Meta object and Passthrough string.
	//
	// See more about [mux direct uploads].
	//
	// [mux direct uploads]: https://www.mux.com/docs/guides/upload-files-directly
	CreateUploadURL(creatorID, title string) (*mux.UploadResponse, error)
	// UpdateMetadata updates mux asset `Meta` object and `Passthrough` string with provided values.
	// All request values are required for update and previous values will be completely deleted.
	UpdateMetadata(req *assetmodel.UpdateMetadataRequest) (*mux.AssetResponse, error)
	// DeleteAsset completely deletes a mux asset. This action is irreversable.
	DeleteAsset(assetID string) error
}

type Client struct {
	client *mux.APIClient
}

type passthroughStruct struct {
	OwnerType string `json:"owner_type"`
}

func NewMUXClient() (MUX, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	muxApiKey := os.Getenv("MUX_API_KEY")
	muxSecretKey := os.Getenv("MUX_SECRET_KEY")
	if muxApiKey == "" || muxSecretKey == "" {
		return nil, fmt.Errorf("MUX_API_KEY or MUX_SECRET_KEY not set in environment")
	}

	client := mux.NewAPIClient(
		mux.NewConfiguration(
			mux.WithBasicAuth(muxApiKey, muxSecretKey),
		),
	)

	return &Client{
		client: client,
	}, nil
}

// CreateUploadURL creates url for direct upload to the mux using mux API.
// It also sets metadata for the created asset using mux assset's Meta object and Passthrough string.
//
// See more about [mux direct uploads].
//
// [mux direct uploads]: https://www.mux.com/docs/guides/upload-files-directly
func (c *Client) CreateUploadURL(creatorID, title string) (*mux.UploadResponse, error) {
	if c.client == nil {
		return nil, fmt.Errorf("mux client is not initialized")
	}
	if _, err := uuid.Parse(creatorID); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidArgument)
	}

	// Define structured metadata object
	assetMeta := mux.AssetMetadata{
		Title:     title,
		CreatorId: creatorID,
	}

	createAssetReq := mux.CreateAssetRequest{
		PlaybackPolicy: []mux.PlaybackPolicy{mux.PUBLIC},
		Meta:           assetMeta,
	}

	createUploadReq := mux.CreateUploadRequest{
		NewAssetSettings: createAssetReq,
		CorsOrigin:       "*",
		Timeout:          3600,
	}

	resp, err := c.client.DirectUploadsApi.CreateDirectUpload(createUploadReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create upload url: %w", ErrAPI, err)
	}

	return &resp, nil
}

// UpdateMetadata updates mux asset `Meta` object and `Passthrough` string with provided values.
// All request values are required for update and previous values will be completely deleted.
func (c *Client) UpdateMetadata(req *assetmodel.UpdateMetadataRequest) (*mux.AssetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	p := passthroughStruct{OwnerType: req.OwnerType}
	passthrough, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal passthrough: %w", err)
	}

	meta := mux.AssetMetadata{
		Title:      req.Title,
		ExternalId: req.OwnerID,
		CreatorId:  req.CreatorID,
	}

	updateReq := mux.UpdateAssetRequest{
		Passthrough: string(passthrough),
		Meta:        meta,
	}

	resp, err := c.client.AssetsApi.UpdateAsset(req.AssetID, updateReq)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update asset metadata: %w", ErrAPI, err)
	}

	return &resp, err
}

// DeleteAsset completely deletes a mux asset. This action is irreversable.
func (c *Client) DeleteAsset(assetID string) error {
	if c.client == nil {
		return fmt.Errorf("mux client is not initialized")
	}

	err := c.client.AssetsApi.DeleteAsset(assetID)
	if err != nil {
		return fmt.Errorf("%w: failed to delete asset: %w", ErrAPI, err)
	}

	return nil
}
