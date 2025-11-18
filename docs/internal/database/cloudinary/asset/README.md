# Cloudinary Asset Repository Documentation

The Cloudinary Asset repository provides data access operations for Cloudinary asset models stored in PostgreSQL. It handles storage and retrieval of asset records including soft-delete functionality for managing the lifecycle of assets in the media service.

## Key Features

- Store and retrieve Cloudinary asset records
- Support for soft-delete and restore operations
- Pagination for listing operations
- Bulk operations on asset records
- Support for transactional operations
- Count operations for analytics

## Dependencies

- **External**: GORM for database interactions with PostgreSQL
- **Internal**: Cloudinary asset models
- **Database**: PostgreSQL with support for GORM operations

## Documentation

- [API Reference](./api.md): Detailed descriptions of repository methods (e.g., `Get`, `List`, `Create`).
- [Architecture](./architecture.md): Technical design and interactions with other components.

## Setup

The repository requires:
- An existing GORM database connection
- Auto-migration setup for the Asset model

## Contributing

To contribute to the Cloudinary Asset repository or its documentation, see [Contributing Guidelines](../../contributing.md).