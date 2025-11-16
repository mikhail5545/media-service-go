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

// Package errors provides custom error handeling and convertions.
package errors

import (
	"errors"

	cloudinaryservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleServiceError(err error) error {
	if errors.Is(err, muxservice.ErrInvalidArgument) ||
		errors.Is(err, cloudinaryservice.ErrInvalidArgument) {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if errors.Is(err, muxservice.ErrNotFound) ||
		errors.Is(err, cloudinaryservice.ErrNotFound) {
		return status.Error(codes.NotFound, err.Error())
	} else if errors.Is(err, cloudinaryservice.ErrInvalidSignature) {
		return status.Error(codes.Unauthenticated, err.Error())
	} else if errors.Is(err, cloudinaryservice.ErrExternalService) {
		return status.Error(codes.Unavailable, err.Error())
	}
	return status.Errorf(codes.Internal, "unexpected error occurred: %s", err.Error())
}
