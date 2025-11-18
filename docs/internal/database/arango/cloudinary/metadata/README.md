# ArangoDB Cloudinary Metadata Repository Documentation

The ArangoDB Cloudinary Metadata repository provides data access operations for Cloudinary asset metadata stored in ArangoDB. It handles storage and retrieval of asset owners and related metadata in the media service.

## Key Features

- Store and retrieve Cloudinary asset metadata
- Manage asset ownership relationships
- Query for unowned assets
- Support for bulk operations on metadata
- Handle metadata collection initialization

## Dependencies

- **External**: ArangoDB Go driver for database interactions
- **Internal**: Cloudinary metadata models, ArangoDB connection package
- **Database**: ArangoDB with "cloudinary_asset_metadata" collection

## Documentation

- [API Reference](./api.md): Detailed descriptions of repository methods (e.g., `Get`, `UpdateOwners`, `DeleteOwners`).
- [Architecture](./architecture.md): Technical design and interactions with other components.

## Setup

The repository requires:
- An existing ArangoDB connection
- The "cloudinary_asset_metadata" collection to exist (or use EnsureCollection to create it)

## Contributing

To contribute to the ArangoDB Cloudinary Metadata repository or its documentation, see [Contributing Guidelines](../../contributing.md).