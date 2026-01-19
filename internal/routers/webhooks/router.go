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

package webhooks

import (
	"github.com/labstack/echo/v4"
	cldhandler "github.com/mikhail5545/media-service-go/internal/handlers/webhooks/cloudinary"
	muxhandler "github.com/mikhail5545/media-service-go/internal/handlers/webhooks/mux"
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
	webhooks := group.Group("/webhooks")

	r.setupCloudinaryRoutes(webhooks)
	r.setupMuxRoutes(webhooks)
}

func (r *RouterImpl) setupCloudinaryRoutes(group *echo.Group) {
	cldGroup := group.Group("/cloudinary")
	handler := cldhandler.New(r.deps.CldSvc)
	cldGroup.POST("", handler.Handle)
}

func (r *RouterImpl) setupMuxRoutes(group *echo.Group) {
	muxGroup := group.Group("/mux")
	handler := muxhandler.New(r.deps.MuxSvc)
	muxGroup.POST("", handler.Handle)
}
