# Media Microservice [![Go Version](https://img.shields.io/badge/go-1.24-blue)](https://golang.org) [![License](https://img.shields.io/badge/license-AGPL-green)](LICENSE)

This microservice manages media assets for the platform, providing APIs for uploading, deleting and updating assets stored in Cloudinary and Mux. It integrates with PostgreSQL for asset records, ArangoDB for metadata, and other services for owner management.

## Features

- Generate signed upload URLs for secure asset uploads to Cloudinary and MUX.
- Delete assets and their associated database records.
- Manage asset ownership and metadata.
- Coordinate with the API Gateway for authentication services.

## Getting Started

### Prerequirements

- Go 1.24 or higher
- PostgreSQL 15
- ArangoDB 3.10
- Cloudinary and MUX API credentials
- Docker (optional, for centralized deployment)

### Installation

1. Clone the repository:

    ```bash
    git clone https://github.com/mikhail5545/media-service-go
    cd media-asset-service
    ```

2. Install dependencies:

    ```bash
    go mod tidy
    ```

3. Set Environment variables:

    ```bash
    export CLOUDINARY_API_KEY=your_key
    export CLOUDINARY_API_SECRET=your_secret
    export CLOUDINARY_CLOUD_NAME=your_cloud_name
    export MUX_TOKEN_ID=your_id
    export MUX_TOKEN_SECRET=your_secret
    export DATABASE_URL=postgres://user:pass@localhost:5432/db
    export ARANGODB_URL=http://localhost:8529
    ```

### Running the Service

Run the service locally:

```bash
go run ./cmd/server
```

Or use Docker:

```bash
docker build -t media-service-go .
docker run -p 8080:8080 media-service-go
```

For detailed setup, see [Development Guide](./docs/development_guide.md).

## Documentation

Detailed documentation is available in the `/docs` directory (see [Table of contents](./docs/table_of_contents.md)):

- Documentation Overview.
- Architecture: System design and high-level processed (e.g., Asset Deletion).
- Cloudinary Service: API and flowcharts for Cloudinary operations.
- Mux Service: API and flowcharts for Mux operations.
- API Reference: Public API endpoints (if applicable).

## License

This project is licensed under the [AGPL-3.0 License](LICENSE.md).
