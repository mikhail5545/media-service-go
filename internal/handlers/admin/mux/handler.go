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

// Package mux provedes HTTP handler admin functionalities for the mux service.
// It acts as an adapter between HTTP transport layer and the underlying service-layer
// mux asset business logic.
package mux

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
	"github.com/mikhail5545/media-service-go/internal/util/request"
)

// Handler holds the service dependency for mux asset-related HTTP handlers.
type Handler struct {
	service muxservice.Service
}

// New creates a new mux handler with the given service.
func New(svc muxservice.Service) *Handler {
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
	if errors.Is(err, muxservice.ErrInvalidArgument) {
		// Log the detailed error for debugging purposes.
		// h.logger.Warn("Invalid argument error", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	} else if errors.Is(err, muxservice.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]any{"error": err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, map[string]any{"error": "Internal server error"})
}

// CreateUploadURL creates upload URL for the direct upload using mux direct upload api. It uses [mux.Client.CreateUploadURL] method
// to access MUX direct upload API. If an owner already has an association with an asset, an error is returned.
//
// Method: POST
// Path: /admin/mux/upload-url
func (h *Handler) CreateUploadURL(c echo.Context) error {
	var req *assetmodel.CreateUploadURLRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	resp, err := h.service.CreateUploadURL(c.Request().Context(), req)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"url": resp.Data.Url})
}

// CreateUnownedUploadURL creates an upload URL for a new asset without an initial owner.
//
// Method: POST
// Path: /admin/mux/upload-url/unowned
func (h *Handler) CreateUnownedUploadURL(c echo.Context) error {
	var req *assetmodel.CreateUnownedUploadURLRequest
	if err := c.Bind(&req); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Invalid request JSON payload")
	}
	resp, err := h.service.CreateUnownedUploadURL(c.Request().Context(), req)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"url": resp.Data.Url})
}

// UpdateOwners processes asset ownership relations changes. It recieves an updated list of asset owners, updates local DB metadata for asset
// (about it's owners), processes the diff between old and new owners and notifies external services about
// this ownership changes via gRPC connection.
//
// Method: PUT
// Path: /admin/mux/assets/:id
func (h *Handler) UpdateOwners(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
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

// Associate handles linking an existing, unowned asset to an owner.
// It expects the asset ID in the URL path and owner information in the JSON body.
//
// Method: POST
// Path: /admin/mux/assets/:id/associate
func (h *Handler) Associate(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
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

// Deassociate handles unlinking an asset from an owner.
// It clears the owner information from the local database and Mux metadata, but does not delete the asset.
//
// Method: POST
// Path: /admin/mux/assets/:id/deassociate
func (h *Handler) Deassociate(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
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

// Get handles retrieving a single, non-deleted asset by its ID.
//
// Method: GET
// Path: /admin/mux/assets/:id
func (h *Handler) Get(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
	if err != nil {
		return err
	}
	response, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"response": response})
}

// GetWithDeleted handles retrieving a single asset by its ID, including soft-deleted ones.
//
// Method: GET
// Path: /admin/mux/assets/:id/with-deleted
func (h *Handler) GetWithDeleted(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
	if err != nil {
		return err
	}
	response, err := h.service.GetWithDeleted(c.Request().Context(), id)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"response": response})
}

// List handles retrieving a paginated list of all non-deleted assets.
// It supports 'limit' and 'offset' query parameters.
//
// Method: GET
// Path: /admin/mux/assets
func (h *Handler) List(c echo.Context) error {
	limit, offset, err := request.GetPaginationParams(c, 10, 0)
	if err != nil {
		return err
	}
	responses, total, err := h.service.List(c.Request().Context(), limit, offset)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"responses": responses, "total": total})
}

// ListUnowned handles retrieving a paginated list of all assets that are not associated with an owner.
// It supports 'limit' and 'offset' query parameters.
//
// Method: GET
// Path: /admin/mux/assets/unowned
func (h *Handler) ListUnowned(c echo.Context) error {
	limit, offset, err := request.GetPaginationParams(c, 10, 0)
	if err != nil {
		return err
	}
	responses, total, err := h.service.ListUnowned(c.Request().Context(), limit, offset)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"responses": responses, "total": total})
}

// ListDeleted handles retrieving a paginated list of all soft-deleted assets.
// It supports 'limit' and 'offset' query parameters.
//
// Method: GET
// Path: /admin/mux/assets/deleted
func (h *Handler) ListDeleted(c echo.Context) error {
	limit, offset, err := request.GetPaginationParams(c, 10, 0)
	if err != nil {
		return err
	}
	responses, total, err := h.service.ListDeleted(c.Request().Context(), limit, offset)
	if err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{"responses": responses, "total": total})
}

// DeletePermanent handles the permanent deletion of an asset from both the local database
// and the Mux service. This action is irreversible.
//
// Method: DELETE
// Path: /admin/mux/assets/:id/permanent
func (h *Handler) DeletePermanent(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
	if err != nil {
		return err
	}
	if err := h.service.DeletePermanent(c.Request().Context(), id); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Delete handles the soft-deletion of an unowned asset.
// The asset is marked as deleted in the database but is not removed from Mux.
//
// Method: DELETE
// Path: /admin/mux/assets/:id
func (h *Handler) Delete(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
	if err != nil {
		return err
	}
	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// Restore handles restoring a soft-deleted asset.
//
// Method: POST
// Path: /admin/mux/assets/:id/restore
func (h *Handler) Restore(c echo.Context) error {
	id, err := request.GetIDParam(c, ":id", "Invalid mux asset ID")
	if err != nil {
		return err
	}
	if err := h.service.Restore(c.Request().Context(), id); err != nil {
		return h.HandleServiceError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}
