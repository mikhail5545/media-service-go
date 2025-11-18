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

// Package cloudinary provedes HTTP handler admin functionalities for the cloudinary service.
// It acts as an adapter between HTTP transport layer and the underlying service-layer
// cloudinary asset business logic.
package cloudinary

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	cloudinaryservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
	"github.com/mikhail5545/media-service-go/internal/util/request"
)

// Handler holds the service dependency for cloudinary asset-related HTTP handlers.
type Handler struct {
	service cloudinaryservice.Service
}

// New creates a new cloudinary handler with the given service.
func New(svc cloudinaryservice.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

// ServeError is a helper function to return a JSON error response.
func (h *Handler) ServeError(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

// HandleServiceError maps service-layer errors to appropriate HTTP status codes.
func (h *Handler) HandleServiceError(c echo.Context, err error) error {
	if errors.Is(err, cloudinaryservice.ErrInvalidArgument) {
		// Log the detailed error for debugging purposes.
		// h.logger.Warn("Invalid argument error", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	} else if errors.Is(err, cloudinaryservice.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
	} else if errors.Is(err, cloudinaryservice.ErrInvalidSignature) {
		return c.JSON(http.StatusForbidden, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
}

// Get retrieves a single not soft-deleted asset record from the database along with it's metadata.
//
// Method: GET
// Path: /admin/mux/assets/:id
func (h *Handler) Get(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	response, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"response": response})
}

// GetWithDeleted retrieves a single asset record from the database along with it's metadata, including soft-deleted ones.
//
// Method: GET
// Path: /admin/mux/assets/deleted/:id
func (h *Handler) GetWithDeleted(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	response, err := h.service.GetWithDeleted(c.Request().Context(), id)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"response": response})
}

// List retrieves a paginated list of all not soft-deleted asset records along with their metadata.
//
// Method: GET
// Path: /admin/mux/assets/
func (h *Handler) List(c echo.Context) error {
	limit, offset, err := request.GetPaginationParams(c, 10, 0)
	if err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid pagination params")
	}
	responses, total, err := h.service.List(c.Request().Context(), limit, offset)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"responses": responses, "total": total})
}

// ListDeleted retrieves a paginated list of all soft-deleted asset records along with their metadata.
//
// Method: GET
// Path: /admin/mux/assets/deleted/
func (h *Handler) ListDeleted(c echo.Context) error {
	limit, offset, err := request.GetPaginationParams(c, 10, 0)
	if err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid pagination params")
	}
	responses, total, err := h.service.ListDeleted(c.Request().Context(), limit, offset)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"responses": responses, "total": total})
}

// ListUnowned retrieves a paginated list of all unowned asset records along with their metadata.
//
// Method: GET
// Path: /admin/mux/assets/unowned/
func (h *Handler) ListUnowned(c echo.Context) error {
	limit, offset, err := request.GetPaginationParams(c, 10, 0)
	if err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid pagination params")
	}
	responses, total, err := h.service.ListUnowned(c.Request().Context(), limit, offset)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"responses": responses, "total": total})
}

// UpdateOwners processes asset ownership relations changes. It recieves an updated list of asset owners, updates local DB metadata for asset
// (about it's owners), processes the diff between old and new owners and notifies external services about this ownership changes via gRPC connection.
//
// Method: PUT
// Path: /admin/mux/assets/:id
func (h *Handler) UpdateOwners(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	var req *assetmodel.UpdateOwnersRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	req.ID = id
	if err := h.service.UpdateOwners(c.Request().Context(), req); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// Associate links an existing asset to an owner. It also updates asset medatada.
//
// Method: POST
// Path: /admin/mux/assets/associate/:id
func (h *Handler) Associate(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	var req *assetmodel.AssociateRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	req.ID = id
	if err := h.service.Associate(c.Request().Context(), req); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// Deassociate removes the link between an asset and an owner. It also deletes owner from asset metadata.
//
// Method: POST
// Path: /admin/mux/assets/deassociate/:id
func (h *Handler) Deassociate(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	var req *assetmodel.DeassociateRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	req.ID = id
	if err := h.service.Deassociate(c.Request().Context(), req); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// SuccessfulUpload creates a new asset with provided information and creates owner relations for it.
// It saves asset metadata about owner relations in the local noSQL db and notifies external services
// about ownership changes via gRPC connection. This method should be called after successful cloudinary image upload.
//
// Method: POST
// Path: /admin/mux/assets/success/:id
func (h *Handler) SuccessfulUpload(c echo.Context) error {
	var req *assetmodel.SuccessfulUploadRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	response, err := h.service.SuccessfulUpload(c.Request().Context(), req)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"response": response})
}

// CreateSignedUploadURL creates a signature for a direct frontend upload.
// Direct upload url should be constructed using this params, this function only creates
// signature for signed upload. It expects a JSON body with Eager, PublicID and File information.
// It returns JSON payload with generated parameters to construct url using them.
// Example: {"signature": "generated_signature", public_id: "asset_public_id", "timestamp": "unix_time", "api_key": "cloudinary_api_key"}.
//
// Method: POST
// Path: /admin/cloudinary/upload-url
func (h *Handler) CreateSignedUploadURL(c echo.Context) error {
	var req *assetmodel.CreateSignedUploadURLRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	res, err := h.service.CreateSignedUploadURL(c.Request().Context(), req)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, map[string]any{"response": res})
}

// Delete performs a soft-delete of an asset. It does not delete Cloudinary asset.
//
// Method: DELETE
// Path: /admin/cloudinary/assets/:id
func (h *Handler) Delete(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// DeletePermanent performs a complete delete of an asset. It also deletes Cloudinary asset.
// This action is irreversable.
//
// Method: DELETE
// Path: /admin/cloudinary/assets/permanent/:id
func (h *Handler) DeletePermanent(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	var req *assetmodel.DestroyAssetRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	req.ID = id
	if err := h.service.DeletePermanent(c.Request().Context(), req); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Restore performs a restore of an asset.
//
// Method: POST
// Path: /admin/cloudinary/assets/restore/:id
func (h *Handler) Restore(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid asset ID")
	if err != nil {
		return err
	}
	if err := h.service.Restore(c.Request().Context(), id); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}
