/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package cloudinary

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikhail5545/media-service-go/internal/handlers/generic"
	cldservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
)

type Handler interface {
	Get(c echo.Context) error
	GetWithArchived(c echo.Context) error
	GetWithBroken(c echo.Context) error
	List(c echo.Context) error
	ListArchived(c echo.Context) error
	ListBroken(c echo.Context) error
	CreateSignedUploadURL(c echo.Context) error
	SuccessfulUpload(c echo.Context) error
	Archive(c echo.Context) error
	Restore(c echo.Context) error
	Delete(c echo.Context) error
	MarkAsBroken(c echo.Context) error
	AddOwner(c echo.Context) error
	RemoveOwner(c echo.Context) error
}

type AdminHandler struct {
	service cldservice.Service
}

var _ Handler = (*AdminHandler)(nil)

func New(svc cldservice.Service) *AdminHandler {
	return &AdminHandler{
		service: svc,
	}
}

func (h *AdminHandler) Get(c echo.Context) error {
	return generic.Handle(c, h.service.Get, http.StatusOK, "asset")
}

func (h *AdminHandler) GetWithArchived(c echo.Context) error {
	return generic.Handle(c, h.service.GetWithArchived, http.StatusOK, "asset")
}

func (h *AdminHandler) GetWithBroken(c echo.Context) error {
	return generic.Handle(c, h.service.GetWithBroken, http.StatusOK, "asset")
}

func (h *AdminHandler) List(c echo.Context) error {
	return generic.HandleList(c, h.service.List, "assets")
}

func (h *AdminHandler) ListArchived(c echo.Context) error {
	return generic.HandleList(c, h.service.List, "assets")
}

func (h *AdminHandler) ListBroken(c echo.Context) error {
	return generic.HandleList(c, h.service.List, "assets")
}

func (h *AdminHandler) CreateSignedUploadURL(c echo.Context) error {
	return generic.Handle(c, h.service.CreateSignedUploadURL, http.StatusOK, "generated")
}

func (h *AdminHandler) SuccessfulUpload(c echo.Context) error {
	return generic.Handle(c, h.service.SuccessfulUpload, http.StatusCreated, "asset")
}

func (h *AdminHandler) Archive(c echo.Context) error {
	return generic.HandleVoid(c, h.service.Archive, http.StatusNoContent)
}

func (h *AdminHandler) Restore(c echo.Context) error {
	return generic.HandleVoid(c, h.service.Restore, http.StatusOK)
}

func (h *AdminHandler) Delete(c echo.Context) error {
	return generic.HandleVoid(c, h.service.Delete, http.StatusNoContent)
}

func (h *AdminHandler) MarkAsBroken(c echo.Context) error {
	return generic.HandleVoid(c, h.service.MarkAsBroken, http.StatusOK)
}

func (h *AdminHandler) AddOwner(c echo.Context) error {
	return generic.HandleVoid(c, h.service.AddOwner, http.StatusCreated)
}

func (h *AdminHandler) RemoveOwner(c echo.Context) error {
	return generic.HandleVoid(c, h.service.RemoveOwner, http.StatusNoContent)
}
