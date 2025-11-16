# Cloudinary Service Documentation

The Cloudinary service handles interactions with Cloudinary’s API for managing assets in the microservice. It provides functionality for creating signed upload URLs, deleting assets, updating owners, and managing metadata stored in a relational database and ArangoDB.

## Key Features

- Generate signed upload URLs for secure asset uploads.
- Delete assets and their associated database records.
- Update asset ownership and metadata.

## Dependencies

- **External**: Cloudinary API for asset management.
- **Internal**: Relational database (PostgreSQL), ArangoDB, and the media service for coordination.
- **Authentication**: Requires JWT tokens verified by the API Gateway.

## Documentation

- [API Reference](./api.md): Detailed descriptions of service methods (e.g., `CreateSignedUploadURL`, `DeleteAsset`).
- [Architecture](./architecture.md): Technical design and interactions with other components.
- [Flowcharts]:
  - [Create Signed Upload URL Flow](../cloudinary/flow/create_signed_upload_url_flow.md): Low-level logic for `CreateSignedUploadURL`.
  - [Update Owners Flow](./update_owners_flow.md): Low-level logic for updating asset owners.
- [High-Level Processes](../../architecture/): Cross-service workflows (e.g., asset deletion).

## Setup

To use the Cloudinary service, configure the following environment variables:

- `CLOUDINARY_API_KEY`: API key for Cloudinary.
- `CLOUDINARY_API_SECRET`: Secret for signing requests.

See [Development Guide](../../guides/development.md) for full setup instructions.

## Contributing

To contribute to the Cloudinary service or its documentation, see [Contributing Guidelines](../../contributing.md).

## Related Documentation

- [Asset Deletion Process](../../architecture/asset_deletion.md)
- [GoDoc](https://pkg.go.dev/your_module/internal/services/cloudinary)