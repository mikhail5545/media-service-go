# PostgreSQL Package API

This document describes the API provided by the PostgreSQL package, which handles PostgreSQL connections in the microservice. The package defines core functionality for establishing connections and initializing the database schema for the Cloudinary asset model.

For an overview of the package, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../architecture/).

## Overview

The PostgreSQL package provides functions to:

- `NewPostgresDB`: Initialize a GORM connection to a PostgreSQL database with auto-migration for the Cloudinary asset model

## NewPostgresDB

The `NewPostgresDB` function initializes and returns a GORM connection to a PostgreSQL database with auto-migration for the Cloudinary asset model.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| dsn       | string          | Required | Database connection string for PostgreSQL |

### Output

| Type      | Description                                    |
|-----------|------------------------------------------------|
| *gorm.DB | GORM database instance if connection successful |
| error     | Error if connection or migration failed, nil otherwise |

### Example

```go
dsn := "host=localhost user=myuser password=mypass dbname=mydb port=5432 sslmode=disable"
db, err := database.NewPostgresDB(context.Background(), dsn)
if err != nil {
    // Handle error
}
// Use db for database operations
```

## Auto-migration

The function automatically performs auto-migration for the `asset.Asset` model, ensuring the database schema is up-to-date with the model definition.