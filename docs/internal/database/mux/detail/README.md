# Mux Detail Repository Documentation

The Mux Detail repository provides data access operations for Mux asset detail models stored in PostgreSQL. It handles storage and retrieval of detailed asset information including tracks and other metadata specific to Mux assets in the media service.

## Key Features

- Store and retrieve Mux asset detail records
- Support for bulk operations on multiple asset details
- Upsert functionality for efficient updates
- Lookup by asset IDs
- Support for transactional operations

## Dependencies

- **External**: GORM for database interactions with PostgreSQL
- **Internal**: Mux asset detail models
- **Database**: PostgreSQL with support for GORM operations

## Documentation

- [API Reference](./api.md): Detailed descriptions of repository methods (e.g., `Get`, `Upsert`, `ListByAssetIDs`).
- [Architecture](./architecture.md): Technical design and interactions with other components.

## Setup

The repository requires:
- An existing GORM database connection
- Auto-migration setup for the Asset Detail model

## Contributing

To contribute to the Mux Detail repository or its documentation, see [Contributing Guidelines](../../contributing.md).