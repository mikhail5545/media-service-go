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

/*
Package cloudinary provides service-layer logic for Cloudinary asset management and asset models.
*/
package cloudinary

import "errors"

var (
	ErrExternalService  = errors.New("external service error")
	ErrNotFound         = errors.New("asset or it's owner not found")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrInvalidSignature = errors.New("invalid request signature")
)
