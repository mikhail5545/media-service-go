// github.com/mikhail5545/media-service-go
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

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/clients/mux"
	"github.com/mikhail5545/media-service-go/internal/clients/productservice"
	"github.com/mikhail5545/media-service-go/internal/database"
	"github.com/mikhail5545/media-service-go/internal/models"
	coursepb "github.com/mikhail5545/proto-go/proto/course/v0"
	muxgo "github.com/muxinc/mux-go"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/gorm"
)

type MuxService struct {
	muxRepo       database.MUXRepository
	muxClient     mux.MUX
	courseService productservice.CourseServiceClient
}

type MUXServiceError struct {
	Msg  string
	Err  error
	Code int
}

func (e *MUXServiceError) Error() string {
	return e.Msg
}

func (e *MUXServiceError) Unwrap() error {
	return e.Err
}

func (e *MUXServiceError) GetCode() int {
	return e.Code
}

func NewMuxService(
	muxRepo database.MUXRepository,
	muxClient mux.MUX,
	courseService productservice.CourseServiceClient,
) *MuxService {
	return &MuxService{
		muxRepo:       muxRepo,
		muxClient:     muxClient,
		courseService: courseService,
	}
}

// CreateCoursePartUploadURL creates upload URL from course part with MUX direct upload API.
// Created asset will include metadata, which contains coures part ID:
//
//	"metadata": {
//		"course_part_id": "ID"
//	}
func (s *MuxService) CreateCoursePartUploadURL(ctx context.Context, partID string) (*muxgo.UploadResponse, error) {
	if _, err := uuid.Parse(partID); err != nil {
		return nil, &MUXServiceError{Msg: "Invalid Course part ID", Err: err, Code: http.StatusBadRequest}
	}

	getResponse, err := s.courseService.GetCoursePart(ctx, &coursepb.GetCoursePartRequest{Id: partID})
	if err != nil {
		return nil, &MUXServiceError{
			Msg:  "Failed to get course part information from course service",
			Err:  err,
			Code: http.StatusServiceUnavailable,
		}
	}

	response, err := s.muxClient.CreateCoursePartUploadURL(partID)
	if err != nil {
		return nil, &MUXServiceError{
			Msg:  "Failed to create upload URL for course part",
			Err:  err,
			Code: http.StatusServiceUnavailable,
		}
	}

	if getResponse.CoursePart.MuxVideoId != nil {
		// If upload already exists, retrieve it from the database
		upload, err := s.GetMuxUpload(ctx, *getResponse.CoursePart.MuxVideoId)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &MUXServiceError{
					Msg:  "MUX Upload not found",
					Err:  err,
					Code: http.StatusNotFound,
				}
			}
			return nil, &MUXServiceError{
				Msg:  "Failed to get mux upload information",
				Err:  err,
				Code: http.StatusServiceUnavailable,
			}
		}

		// Delete old Asset with MUX API
		if upload.MUXAssetID != nil {
			if err := s.muxClient.DeleteMUXAsset(*upload.MUXAssetID); err != nil {
				return nil, &MUXServiceError{
					Msg:  "Failed to delete old mux asset",
					Err:  err,
					Code: http.StatusServiceUnavailable,
				}
			}
		}

		// Create new asset and upload URL for it if old Asset was successfully deleted
		response, err = s.muxClient.CreateCoursePartUploadURL(partID)
		if err != nil {
			return nil, &MUXServiceError{
				Msg:  "Failed to create upload URL for course part",
				Err:  err,
				Code: http.StatusServiceUnavailable,
			}
		}

		// Populate and update mux upload instance with new data, no changes in Course Part record needed
		upload.VideoProcessingStatus = "upload_url_created"
		upload.MUXUploadID = &response.Data.Id
		upload.MUXAssetID = &response.Data.NewAssetSettings.Id
		if err := s.muxRepo.Update(ctx, upload); err != nil {
			return nil, &MUXServiceError{
				Msg:  "Failed to update mux upload",
				Err:  err,
				Code: http.StatusInternalServerError,
			}
		}
	} else {
		upload, err := s.CreateMuxUpload(ctx, response.Data.Id, "upload_url_created", getResponse.CoursePart.Id)
		if err != nil {
			return nil, &MUXServiceError{
				Msg:  "Failed to create mux upload",
				Err:  err,
				Code: http.StatusInternalServerError,
			}
		}

		_, err = s.courseService.AddMuxVideoToCoursePart(ctx, &coursepb.AddMuxVideoToCoursePartRequest{
			Id:         partID,
			MuxVideoId: upload.ID, // MUXUpload ID (uuid string, not MUX API direct upload ID)
		})
		if err != nil {
			return nil, &MUXServiceError{
				Msg:  "Failed to add mux video to course part via course service",
				Err:  err,
				Code: http.StatusServiceUnavailable,
			}
		}
	}

	return &response, nil
}

func (s *MuxService) GetMuxUpload(ctx context.Context, id string) (*models.MUXUpload, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, &MUXServiceError{Msg: "Invalid Mux Upload ID", Err: err, Code: http.StatusBadRequest}
	}

	upload, err := s.muxRepo.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &MUXServiceError{Msg: "Mux Upload not found", Err: err, Code: http.StatusNotFound}
		}
		return nil, &MUXServiceError{Msg: "Failed to find Mux Upload", Err: err, Code: http.StatusInternalServerError}
	}

	return upload, nil
}

// CreateMuxUpload creates new MUXUpload record in the database.
// It calls for github.com/mikhail5545/product-service-go to retrieve CoursePart and update it
// with newly created MUXUpload information.
func (s *MuxService) CreateMuxUpload(ctx context.Context, uploadID string, status string, partID string) (*models.MUXUpload, error) {
	var muxVideo models.MUXUpload
	err := s.muxRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMuxRepo := s.muxRepo.WithTx(tx)

		// Retrieve CoursePart record from product-service-go via gRPC connection.
		getResponse, err := s.courseService.GetCoursePart(ctx, &coursepb.GetCoursePartRequest{
			Id: partID,
		})
		if err != nil {
			return &MUXServiceError{
				Msg:  "Failed to get course part from course service",
				Err:  err,
				Code: http.StatusServiceUnavailable,
			}
		}

		// Check if some MUXUpload already binded to the course part.
		if getResponse.CoursePart.MuxVideo != nil {
			return &MUXServiceError{
				Msg:  "MUXVideo instance already exists for this part",
				Err:  fmt.Errorf("MUXVideo instance already exists for this part"),
				Code: http.StatusBadRequest,
			}
		}

		muxVideo = models.MUXUpload{
			ID:                    uuid.New().String(),
			MUXUploadID:           &uploadID, // MUX API direct upload id
			VideoProcessingStatus: status,
		}

		updateReq := coursepb.UpdateCoursePartRequest{
			MuxVideoId: &muxVideo.ID,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"mux_video_id"}, // Which fields client (this service) wants to update.
			},
		}

		// Update CoursePart in the product-service-go
		_, err = s.courseService.UpdateCoursePart(ctx, &updateReq)
		if err != nil {
			return &MUXServiceError{
				Msg:  "Failed to get course part via course service",
				Err:  err,
				Code: http.StatusServiceUnavailable,
			}
		}

		if err := txMuxRepo.Create(ctx, &muxVideo); err != nil {
			return &MUXServiceError{
				Msg:  "Failed to create mux video",
				Err:  err,
				Code: http.StatusInternalServerError,
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &muxVideo, nil
}

func (s *MuxService) UpdateMuxUpload(ctx context.Context, id string, upload *models.MUXUpload, fieldMask *fieldmaskpb.FieldMask) (*models.MUXUpload, error) {
	var uploadToUpdate *models.MUXUpload
	err := s.muxRepo.DB().Transaction(func(tx *gorm.DB) error {
		if _, err := uuid.Parse(id); err != nil {
			return &MUXServiceError{
				Msg:  "Invalid Mux Upload ID",
				Err:  err,
				Code: http.StatusBadRequest,
			}
		}
		txMuxRepo := s.muxRepo.WithTx(tx)

		if fieldMask != nil {

		}
		var findErr error
		uploadToUpdate, findErr = txMuxRepo.Find(ctx, id)
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return &MUXServiceError{
					Msg:  "MUX Upload not found",
					Err:  findErr,
					Code: http.StatusNotFound,
				}
			}
			return &MUXServiceError{
				Msg:  "Failed to get mux upload information",
				Err:  findErr,
				Code: http.StatusServiceUnavailable,
			}
		}

		var updated bool
		// This field cannot be null in case of update
		if upload.MUXUploadID != nil && *upload.MUXUploadID != *uploadToUpdate.MUXUploadID {
			uploadToUpdate.MUXUploadID = upload.MUXUploadID
			updated = true
		}
		// This field cannot be null in case of update
		if upload.MUXAssetID != nil && *upload.MUXAssetID != *uploadToUpdate.MUXAssetID {
			uploadToUpdate.MUXAssetID = upload.MUXAssetID
			updated = true
		}
		if *upload.MUXPlaybackID != *uploadToUpdate.MUXPlaybackID {
			uploadToUpdate.MUXPlaybackID = upload.MUXPlaybackID
			updated = true
		}
		// This field cannot be blank in case of update
		if upload.VideoProcessingStatus != "" && upload.VideoProcessingStatus != uploadToUpdate.VideoProcessingStatus {
			uploadToUpdate.VideoProcessingStatus = upload.VideoProcessingStatus
			updated = true
		}
		if *upload.Duration != *uploadToUpdate.Duration {
			uploadToUpdate.Duration = upload.Duration
			updated = true
		}
		if *upload.AspectRatio != *uploadToUpdate.AspectRatio {
			uploadToUpdate.AspectRatio = upload.AspectRatio
			updated = true
		}
		if *upload.MaxHeight != *uploadToUpdate.MaxHeight {
			uploadToUpdate.MaxHeight = upload.MaxHeight
			updated = true
		}
		if *upload.MaxWidth != *uploadToUpdate.MaxWidth {
			uploadToUpdate.MaxWidth = upload.MaxWidth
			updated = true
		}
		if *upload.AssetCreatedAt != *uploadToUpdate.AssetCreatedAt {
			uploadToUpdate.AssetCreatedAt = upload.AssetCreatedAt
			updated = true
		}

		if updated {
			if err := txMuxRepo.Update(ctx, uploadToUpdate); err != nil {
				return &MUXServiceError{
					Msg:  "Failed to update MUX upload",
					Err:  err,
					Code: http.StatusInternalServerError,
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return uploadToUpdate, nil
}

func (s *MuxService) DeleteMuxUpload(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return &MUXServiceError{
			Msg:  "Invalid Mux Upload ID",
			Err:  err,
			Code: http.StatusBadRequest,
		}
	}

	err := s.muxClient.DeleteMUXAsset(id)
	if err != nil {
		return &MUXServiceError{
			Msg:  "Failed to create MUX asset",
			Err:  err,
			Code: http.StatusServiceUnavailable,
		}
	}
	if err := s.muxRepo.Delete(ctx, id); err != nil {
		return &MUXServiceError{
			Msg:  "Failed to delete MUX Upload",
			Err:  err,
			Code: http.StatusInternalServerError,
		}
	}
	return nil
}
