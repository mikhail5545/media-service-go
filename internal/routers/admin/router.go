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

package admin

import (
	"github.com/labstack/echo/v4"
	cldhandler "github.com/mikhail5545/media-service-go/internal/handlers/admin/cloudinary"
	muxhandler "github.com/mikhail5545/media-service-go/internal/handlers/admin/mux"
	"github.com/mikhail5545/media-service-go/internal/routers"
	cldservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
)

type Dependencies struct {
	MuxSvc *muxservice.Service
	CldSvc *cldservice.Service
}

type RouterImpl struct {
	deps Dependencies
}

var _ routers.Router = (*RouterImpl)(nil)

func New(d Dependencies) *RouterImpl {
	return &RouterImpl{deps: d}
}

func (r *RouterImpl) Setup(group *echo.Group) {
	admin := group.Group("/admin")

	r.setupHealthRoutes(admin)
	r.setupMuxRoutes(admin)
	r.setupCloudinaryRoutes(admin)
}

func (r *RouterImpl) setupHealthRoutes(group *echo.Group) {
	group.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
}

func (r *RouterImpl) setupMuxRoutes(group *echo.Group) {
	handler := muxhandler.New(r.deps.MuxSvc)

	muxGroup := group.Group("/mux")
	{
		assets := muxGroup.Group("/assets")
		{
			assets.GET("/:id", handler.Get)
			assets.GET("/archived/:id", handler.GetWithArchived)
			assets.GET("/broken/:id", handler.GetWithBroken)
			assets.GET("", handler.List)
			assets.GET("/archived", handler.ListArchived)
			assets.GET("/broken", handler.ListBroken)
			assets.POST("/upload-url", handler.CreateUploadURL)
			assets.DELETE("/archive/:id", handler.Archive)
			assets.POST("/restore/:id", handler.Restore)
			assets.DELETE("/:id", handler.Delete)
			assets.POST("/broken/:id", handler.MarkAsBroken)
			assets.POST("/:id/owners", handler.AddOwner)
			assets.DELETE("/:id/owners", handler.RemoveOwner)
		}
	}
}

func (r *RouterImpl) setupCloudinaryRoutes(group *echo.Group) {
	handler := cldhandler.New(r.deps.CldSvc)

	cldGroup := group.Group("/cloudinary")
	{
		assets := cldGroup.Group("/assets")
		{
			assets.GET("/:id", handler.Get)
			assets.GET("/archived/:id", handler.GetWithArchived)
			assets.GET("/broken/:id", handler.GetWithBroken)
			assets.GET("", handler.List)
			assets.GET("/archived", handler.ListArchived)
			assets.GET("/broken", handler.ListBroken)
			assets.POST("/upload/url-gen", handler.CreateSignedUploadURL)
			assets.POST("/upload/success", handler.SuccessfulUpload)
			assets.DELETE("/archive/:id", handler.Archive)
			assets.POST("/restore/:id", handler.Restore)
			assets.DELETE("/:id", handler.Delete)
			assets.POST("/broken/:id", handler.MarkAsBroken)
			assets.POST("/:id/owners", handler.AddOwner)
			assets.DELETE("/:id/owners", handler.RemoveOwner)
		}
	}
}
