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

package cloudinary

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/joho/godotenv"
)

type Cloudinary interface {
	DeleteAsset(ctx context.Context, publicID string, assetType string) error
}

type CloudinaryClient struct {
	client *cloudinary.Cloudinary
}

type CloudinaryError struct {
	Msg  string
	Err  error
	Code int
}

func (e *CloudinaryError) Error() string {
	return e.Msg
}

func (e *CloudinaryError) Unwrap() error {
	return e.Err
}

func (e *CloudinaryError) GetCode() int {
	return e.Code
}

func NewCloudinaryClient() (Cloudinary, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, &CloudinaryError{
			Msg:  "missing required environment variables",
			Err:  fmt.Errorf("enfironment variables missing"),
			Code: http.StatusInternalServerError,
		}
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, &CloudinaryError{
			Msg:  "Failed to initialize cloudinary client",
			Err:  err,
			Code: http.StatusInternalServerError,
		}
	}
	return &CloudinaryClient{client: cld}, nil
}

func (c *CloudinaryClient) DeleteAsset(ctx context.Context, publicID string, assetType string) error {
	if publicID == "" {
		return &CloudinaryError{
			Msg:  "assetType is required",
			Err:  fmt.Errorf("missing parameters"),
			Code: http.StatusBadRequest,
		}
	}
	if assetType == "" {
		return &CloudinaryError{
			Msg:  "assetType is required",
			Err:  fmt.Errorf("missing parameters"),
			Code: http.StatusBadRequest,
		}
	}

	res, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: assetType,
	})
	if err != nil {
		return &CloudinaryError{
			Msg:  fmt.Sprintf("Failed to delete cloudinary asset: %s", res.Response),
			Err:  fmt.Errorf("%v; result: %v", err, res.Error),
			Code: http.StatusServiceUnavailable,
		}
	}

	return nil
}
