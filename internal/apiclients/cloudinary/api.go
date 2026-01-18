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

package cloudinary

import (
	"context"
	"fmt"
	"net/url"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type APIClient interface {
	SignUploadParams(ctx context.Context, params url.Values) (string, error)
	VerifyNotificationSignature(ctx context.Context, params *VerificationParams) bool
	GetApiKey() string
}

type Client struct {
	client *cloudinary.Cloudinary
}

var _ APIClient = (*Client)(nil)

func New(cloudName, apiKey, apiSecret string) (*Client, error) {
	if cloudName == "" {
		return nil, fmt.Errorf("cloud name is required")
	}
	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("api key and secret are required")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloudinary client: %w", err)
	}

	return &Client{
		client: cld,
	}, nil
}

func (c *Client) SignUploadParams(ctx context.Context, params url.Values) (string, error) {
	signature, err := api.SignParameters(params, c.client.Config.Cloud.APISecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign upload parameters: %w", err)
	}
	return signature, nil
}

type VerificationParams struct {
	Payload           string
	ReceivedSignature string
	Timestamp         int64
	ValidFor          int64
}

func (c *Client) VerifyNotificationSignature(ctx context.Context, params *VerificationParams) bool {
	// NOTE: Cloudinary sets `validFor` as two hours if `validFor` is 0 by default.
	return c.client.Upload.VerifyNotificationSignature(params.Payload, params.Timestamp, params.ReceivedSignature, params.ValidFor)
}

func (c *Client) DeleteAsset(ctx context.Context, publicID string, resourceType string) error {
	if publicID == "" {
		return fmt.Errorf("publicID is required")
	}
	if resourceType == "" {
		return fmt.Errorf("resourceType is required")
	}

	_, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	})

	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}

func (c *Client) DeleteAssets(ctx context.Context, assetType string, publicIDs []string) error {
	ids := api.CldAPIArray{}
	ids = append(ids, publicIDs...)

	if assetType == "" {
		return fmt.Errorf("assetType is required")
	}
	if len(publicIDs) > 100 {
		return fmt.Errorf("public ids length cannot be greater that 100")
	}

	_, err := c.client.Admin.DeleteAssets(ctx, admin.DeleteAssetsParams{
		AssetType: api.AssetType(assetType),
		PublicIDs: ids,
	})
	if err != nil {
		return fmt.Errorf("failed to delete assets %w", err)
	}
	return nil
}

func (c *Client) CreateFolder(ctx context.Context, folder string) (bool, error) {
	if folder == "" {
		return false, fmt.Errorf("folder is required")
	}
	res, err := c.client.Admin.CreateFolder(ctx, admin.CreateFolderParams{Folder: folder})
	if err != nil {
		return false, fmt.Errorf("failed to create folder: %w", err)
	}
	return res.Success, nil
}

func (c *Client) GetRootFolders(ctx context.Context, maxResults int) (*admin.FoldersResult, error) {
	res, err := c.client.Admin.RootFolders(ctx, admin.RootFoldersParams{MaxResults: maxResults})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve root folders: %w", err)
	}
	return res, nil
}

func (c *Client) ListAssetsByFolder(ctx context.Context, folder string) ([]api.BriefAssetResult, error) {
	if folder == "" {
		return nil, fmt.Errorf("folder is required")
	}
	res, err := c.client.Admin.AssetsByAssetFolder(ctx, admin.AssetsByAssetFolderParams{
		AssetFolder: folder,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve assets by folder: %w", err)
	}
	return res.Assets, nil
}

func (c *Client) GetApiKey() string {
	return c.client.Config.Cloud.APIKey
}
