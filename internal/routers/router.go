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

package routers

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	admincloudinaryhandler "github.com/mikhail5545/media-service-go/internal/handlers/admin/cloudinary"
	adminmuxhandler "github.com/mikhail5545/media-service-go/internal/handlers/admin/mux"
	cldwebhookhandler "github.com/mikhail5545/media-service-go/internal/handlers/hooks/cloudinary"
	muxwebhookhandler "github.com/mikhail5545/media-service-go/internal/handlers/hooks/mux"
	cloudinaryservice "github.com/mikhail5545/media-service-go/internal/services_outdated/cloudinary"
	muxservice "github.com/mikhail5545/media-service-go/internal/services_outdated/mux"
)

func SetupRouter(e *echo.Echo, muxService muxservice.Service, cldService cloudinaryservice.Service) {
	api := e.Group("/api")
	ver := api.Group("/v0")

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// --- Admin handlers ---
	muxAdminHandler := adminmuxhandler.New(muxService)
	cldAdminHandler := admincloudinaryhandler.New(cldService)

	// --- Webhook handlers ---
	muxWebhookHandler := muxwebhookhandler.New(muxService)
	cldWebhookHandler := cldwebhookhandler.New(cldService)

	admin := ver.Group("/admin")
	{
		adminMux := admin.Group("/mux")
		{
			adminMux.POST("/upload-url", muxAdminHandler.CreateUploadURL)
			adminMux.POST("/upload-url/unowned", muxAdminHandler.CreateUnownedUploadURL)

			assets := adminMux.Group("/assets")
			{
				assets.POST("/associate/:id", muxAdminHandler.Associate)
				assets.POST("/deassociate/:id", muxAdminHandler.Deassociate)
				assets.GET("", muxAdminHandler.List)
				assets.GET("/unowned", muxAdminHandler.ListUnowned)
				assets.GET("/deleted", muxAdminHandler.ListDeleted)
				assets.GET("/:id", muxAdminHandler.Get)
				assets.PUT("/:id", muxAdminHandler.UpdateOwners)
				assets.GET("/deleted/:id", muxAdminHandler.GetWithDeleted)
				assets.DELETE("/:id", muxAdminHandler.Delete)
				assets.DELETE("/permanent/:id", muxAdminHandler.DeletePermanent)
				assets.POST("/restore/:id", muxAdminHandler.Restore)
			}
		}

		adminCld := admin.Group("/cloudinary")
		{
			adminCld.POST("/upload-url", cldAdminHandler.CreateSignedUploadURL)

			assets := adminCld.Group("/assets")
			{
				assets.GET("", cldAdminHandler.List)
				assets.GET("/deleted", cldAdminHandler.ListDeleted)
				assets.GET("/unowned", cldAdminHandler.ListUnowned)
				assets.GET("/:id", cldAdminHandler.Get)
				assets.GET("/deleted/:id", cldAdminHandler.GetWithDeleted)
				assets.POST("/associate/:id", cldAdminHandler.Associate)
				assets.POST("/deassocaite/:id", cldAdminHandler.Deassociate)
				assets.PUT("/:id", cldAdminHandler.UpdateOwners)
				assets.DELETE("/:id", cldAdminHandler.Delete)
				assets.DELETE("/permanent/:id", cldAdminHandler.DeletePermanent)
				assets.POST("/restore/:id", cldAdminHandler.Restore)
			}
		}
	}

	webhooks := ver.Group("/webhooks")
	{
		mux := webhooks.Group("/mux")
		{
			mux.POST("", muxWebhookHandler.HandleWebhook)
		}

		cloudinary := webhooks.Group("/cloudinary")
		{
			cloudinary.POST("/upload", cldWebhookHandler.UploadWebhook)
		}
	}
}
