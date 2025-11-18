# ArangoDB Mux Metadata Repository Documentation

The ArangoDB Mux Metadata repository provides data access operations for Mux asset metadata stored in ArangoDB. It handles storage and retrieval of asset owners, titles, and other related metadata in the media service.

## Key Features

- Store and retrieve Mux asset metadata
- Manage asset ownership relationships
- Handle partial updates of metadata fields
- Query for unowned assets
- Support for bulk operations on metadata
- Handle metadata collection initialization

## Dependencies

- **External**: ArangoDB Go driver for database interactions
- **Internal**: Mux metadata models, ArangoDB connection package
- **Database**: ArangoDB with "mux_asset_metadata" collection

## Documentation

- [API Reference](./api.md): Detailed descriptions of repository methods (e.g., `Get`, `Update`, `Create`).
- [Architecture](./architecture.md): Technical design and interactions with other components.

## Setup

The repository requires:
- An existing ArangoDB connection
- The "mux_asset_metadata" collection to exist (or use EnsureCollection to create it)

## Contributing

To contribute to the ArangoDB Mux Metadata repository or its documentation, see [Contributing Guidelines](../../contributing.md).