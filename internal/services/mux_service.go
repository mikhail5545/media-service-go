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

func (s *MuxService) CreateMuxUpload(ctx context.Context, uploadID string, status string, partID string) (*models.MUXUpload, error) {
	var muxVideo models.MUXUpload
	err := s.muxRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMuxRepo := s.muxRepo.WithTx(tx)

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

		if getResponse.CoursePart.MuxVideo != nil {
			return &MUXServiceError{
				Msg:  "MUXVideo instance already exists for this part",
				Err:  fmt.Errorf("MUXVideo instance already exists for this part"),
				Code: http.StatusBadRequest,
			}
		}

		muxVideo = models.MUXUpload{
			ID:                    uuid.New().String(),
			MUXUploadID:           &uploadID,
			VideoProcessingStatus: status,
		}

		updateReq := coursepb.UpdateCoursePartRequest{
			MuxVideoId: &muxVideo.ID,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"mux_video_id"},
			},
		}

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
