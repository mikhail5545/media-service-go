package postgres

import (
	"context"

	cldassetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	muxassetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.WithContext(ctx).AutoMigrate(
		&muxassetmodel.Asset{},
		&cldassetmodel.Asset{},
	)
	if err != nil {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}
