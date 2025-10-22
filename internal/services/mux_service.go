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
	"net/http"

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/clients/mux"
	"github.com/mikhail5545/media-service-go/internal/database"
	"github.com/mikhail5545/media-service-go/internal/models"
	"gorm.io/gorm"
)

type MuxService struct {
	muxRepo   database.MUXRepository
	muxClient mux.MUX
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

func NewMuxService(muxRepo database.MUXRepository, muxClient mux.MUX) *MuxService {
	return &MuxService{
		muxRepo:   muxRepo,
		muxClient: muxClient,
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
