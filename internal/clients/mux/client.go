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

	"github.com/joho/godotenv"
	mux "github.com/muxinc/mux-go"
)

type MUX interface {
	// CreateCoursePartUploadURL creates url to interact with mux API for course part model
	// https://www.mux.com/docs/guides/upload-files-directly
	CreateCoursePartUploadURL(coursePartID string) (mux.UploadResponse, error)
	DeleteMUXAsset(assetID string) error
}

type MUXClient struct {
	client *mux.APIClient
}

type metadata struct {
	CoursePartID string `json:"course_part_id"`
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

	return &MUXClient{
		client: client,
	}, nil
}

// CreateCoursePartUploadURL creates url to interact with mux API for course part model
// https://www.mux.com/docs/guides/upload-files-directly
func (c *MUXClient) CreateCoursePartUploadURL(coursePartID string) (mux.UploadResponse, error) {
	if c.client == nil {
		return mux.UploadResponse{}, fmt.Errorf("mux client is not initialized")
	}

	m := metadata{CoursePartID: coursePartID}
	passthrough, err := json.Marshal(m)
	if err != nil {
		return mux.UploadResponse{}, fmt.Errorf("failed to marshal metadata: %s", err.Error())
	}

	car := mux.CreateAssetRequest{PlaybackPolicy: []mux.PlaybackPolicy{mux.PUBLIC}, Passthrough: string(passthrough)}
	cur := mux.CreateUploadRequest{NewAssetSettings: car, Timeout: 3600, CorsOrigin: "*"}
	u, err := c.client.DirectUploadsApi.CreateDirectUpload(cur)
	if err != nil {
		return mux.UploadResponse{}, fmt.Errorf("failed to create upload url for mux API: %s", err.Error())
	}

	return u, nil
}

func (c *MUXClient) DeleteMUXAsset(assetID string) error {
	if c.client == nil {
		return fmt.Errorf("mux client is not initialized")
	}

	err := c.client.AssetsApi.DeleteAsset(assetID)
	if err != nil {
		return fmt.Errorf("failed to delete asset from mux API: %s", err.Error())
	}

	return nil
}
