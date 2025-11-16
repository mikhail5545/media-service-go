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

// Package cloudinary provides cloudinaruy API client implementation.
package cloudinary

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/admin"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var (
	// ErrInvalidArgument invalid argument error
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrCloudinaryAPI cloudinary api error
	ErrCloudinaryAPI = errors.New("cloudinary api error")
)

// Cloudinary provides cloudinary API client logic.
type Cloudinary interface {
	// DeleteAssets deletes cloudinary assets by they're publicID.
	// Can delete up to 100 assets.
	//
	// Returns an error if the assetType is missing or publicIDs > 100 (ErrInvalidArgument) or
	// Cloudinary API error occures (ErrCloudinaryAPI).
	DeleteAssets(ctx context.Context, assetType string, publicIDs []string) error
	// DeleteAsset deletes cloudinary asset by it's publicID.
	//
	// Returns an error if the assetType or resourceType is missing or(ErrInvalidArgument) or
	// Cloudinary API error occures (ErrCloudinaryAPI).
	DeleteAsset(ctx context.Context, publicID string, assetType string) error
	// CreateFolder creates a new folder.
	//
	// Returns an error if the folder is missing (ErrInvalidArgument) or a Cloudinary API error occures (ErrCloudinaryAPI).
	CreateFolder(ctx context.Context, folder string) (bool, error)
	// GetRootFolders returns a list of all folders in the root directory of the cloudinary cloud.
	//
	// Returns an error if Cloudinary API error occures (ErrCloudinaryAPI).
	GetRootFolders(ctx context.Context, maxResults int) (*admin.FoldersResult, error)
	// SignUploadParams creates a signature for provided upload params.
	//
	// Returns an error if Cloudinary API error occures (ErrCloudinaryAPI).
	SignUploadParams(ctx context.Context, params url.Values) (string, error)
	// VerifyNotificationSignature verifies `recievedSignature`.
	VerifyNotificationSignature(ctx context.Context, payload, recievedSignature string, timestamp, validFor int64) bool
	// ListAssetsByFolder returns a list of all assets located in the specified folder.
	//
	// Returns an error if folder is missing (ErrInvalidArgument) or a Cloudinary API error occures (ErrCloudinaryAPI).
	ListAssetsByFolder(ctx context.Context, folder string) ([]api.BriefAssetResult, error)
	// GetApiKey returns cloudinary API cloud api_key.
	GetApiKey() string
}

// Client implements cloudinary API client logic and holds cloudinary api client instance to perform api calls.
type Client struct {
	client *cloudinary.Cloudinary
}

// NewCloudinaryClient creates a new instance of Client.
// It initialized cloudinary api client using environment variables. If they're not set, error will be returned.
func NewCloudinaryClient() (Cloudinary, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing 'CLOUDINARY_CLOUD_NAME', 'CLOUDINARY_API_KEY', 'CLOUDINARY_API_SECRET' environment variables")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		// This part of your original code was fine, but for completeness:
		// It's better to load .env at the start of main()
		if _, ok := err.(*os.PathError); ok {
			log.Println("Warning: .env file not found. Relying on environment variables.")
		} else {
			log.Fatal("Error loading .env file")
		}

		return nil, fmt.Errorf("failed to initialize cloudinary api client: %w", err)
	}
	return &Client{client: cld}, nil
}

// DeleteAssets deletes cloudinary assets by they're publicID.
// Can delete up to 100 assets.
//
// Returns an error if the assetType is missing or publicIDs > 100 (ErrInvalidArgument) or
// Cloudinary API error occures (ErrCloudinaryAPI).
func (c *Client) DeleteAssets(ctx context.Context, assetType string, publicIDs []string) error {
	ids := api.CldAPIArray{}
	ids = append(ids, publicIDs...)

	if assetType == "" {
		return fmt.Errorf("%w: assetType is required", ErrInvalidArgument)
	}
	if len(publicIDs) > 100 {
		return fmt.Errorf("%w: public ids length cannot be creater that 100", ErrInvalidArgument)
	}

	_, err := c.client.Admin.DeleteAssets(ctx, admin.DeleteAssetsParams{
		AssetType: api.AssetType(assetType),
		PublicIDs: ids,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCloudinaryAPI, err)
	}
	return nil
}

// DeleteAsset deletes cloudinary asset by it's publicID.
//
// Returns an error if the assetType or resourceType is missing or(ErrInvalidArgument) or
// Cloudinary API error occures (ErrCloudinaryAPI).
func (c *Client) DeleteAsset(ctx context.Context, publicID string, resourceType string) error {
	if publicID == "" {
		return fmt.Errorf("%w: publicID is required", ErrInvalidArgument)
	}
	if resourceType == "" {
		return fmt.Errorf("%w: resourceType is required", ErrInvalidArgument)
	}

	_, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: resourceType,
	})

	if err != nil {
		return fmt.Errorf("%w: %w", ErrCloudinaryAPI, err)
	}

	return nil
}

// VerifyNotificationSignature verifies `recievedSignature`.
func (c *Client) VerifyNotificationSignature(ctx context.Context, payload, recievedSignature string, timestamp, validFor int64) bool {
	// NOTE: Cloudinary sets `validFor` as two hours if `validFor` is 0 by default.
	return c.client.Upload.VerifyNotificationSignature(payload, timestamp, recievedSignature, validFor)
}

// CreateFolder creates a new folder.
//
// Returns an error if the folder is missing (ErrInvalidArgument) or a Cloudinary API error occures (ErrCloudinaryAPI).
func (c *Client) CreateFolder(ctx context.Context, folder string) (bool, error) {
	if folder == "" {
		return false, fmt.Errorf("%w: folder is required", ErrInvalidArgument)
	}
	res, err := c.client.Admin.CreateFolder(ctx, admin.CreateFolderParams{Folder: folder})
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrCloudinaryAPI, err)
	}
	return res.Success, nil
}

// GetRootFolders returns a list of all folders in the root directory of the cloudinary cloud.
//
// Returns an error if Cloudinary API error occures (ErrCloudinaryAPI).
func (c *Client) GetRootFolders(ctx context.Context, maxResults int) (*admin.FoldersResult, error) {
	res, err := c.client.Admin.RootFolders(ctx, admin.RootFoldersParams{MaxResults: maxResults})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCloudinaryAPI, err)
	}
	return res, nil
}

// ListAssetsByFolder returns a list of all assets located in the specified folder.
//
// Returns an error if folder is missing (ErrInvalidArgument) or a Cloudinary API error occures (ErrCloudinaryAPI).
func (c *Client) ListAssetsByFolder(ctx context.Context, folder string) ([]api.BriefAssetResult, error) {
	if folder == "" {
		return nil, fmt.Errorf("%w: folder is required", ErrInvalidArgument)
	}
	res, err := c.client.Admin.AssetsByAssetFolder(ctx, admin.AssetsByAssetFolderParams{
		AssetFolder: folder,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCloudinaryAPI, err)
	}
	return res.Assets, nil
}

// SignUploadParams creates a signature for provided upload params.
//
// Returns an error if Cloudinary API error occures (ErrCloudinaryAPI).
func (c *Client) SignUploadParams(ctx context.Context, params url.Values) (string, error) {
	signature, err := api.SignParameters(params, c.client.Config.Cloud.APIKey)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrCloudinaryAPI, err)
	}

	return signature, nil
}

// GetApiKey returns cloudinary API cloud api_key.
func (c *Client) GetApiKey() string {
	return c.client.Config.Cloud.APIKey
}
