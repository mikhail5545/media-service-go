# PostgreSQL Database Package Documentation

The PostgreSQL package provides database connectivity and initialization functionality for PostgreSQL in the media service. It handles GORM initialization, connection setup, and model auto-migration for the Cloudinary asset model.

## Key Features

- Initialize GORM connection to PostgreSQL
- Support for PostgreSQL-specific configuration
- Auto-migration for Cloudinary asset models
- Context-aware database operations
- Error handling for connection and migration failures

## Dependencies

- **External**: GORM with PostgreSQL driver for database interactions
- **Internal**: Cloudinary asset models
- **Database**: PostgreSQL server with required schema permissions

## Documentation

- [API Reference](./api.md): Detailed descriptions of package functions (e.g., `NewPostgresDB`).
- [Architecture](./architecture.md): Technical design and interactions with other components.

## Setup

To use the PostgreSQL package, ensure access to a PostgreSQL database with appropriate permissions for:
- Creating tables (for auto-migration)
- Reading and writing to the assets table

## Contributing

To contribute to the PostgreSQL package or its documentation, see [Contributing Guidelines](../../contributing.md).