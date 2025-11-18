# Cloudinary Asset Repository Architecture

This document describes the technical design of the Cloudinary Asset repository, which manages Cloudinary asset data within the microservice.

## Repository Architecture

The Cloudinary Asset repository is defined by the `Repository` interface and implemented by the `gormRepository` struct. It consists of:

- **Data Access Layer**: Provides methods for interacting with the PostgreSQL database using GORM
- **Soft Delete Support**: Implements soft deletion using GORM's built-in functionality
- **Transaction Management**: Supports database transactions through the WithTx method
- **Query Management**: Handles various query patterns including pagination and filtering
- **Error Handling**: Uses GORM's error handling patterns

The repository follows a standard repository pattern architecture:

1. **Interface Definition**: Defines the contract for asset operations
2. **GORM Implementation**: Provides concrete PostgreSQL-based implementation using GORM
3. **Query Execution**: Executes database queries using GORM's query builder
4. **Result Handling**: Processes and returns results

## Interactions

- **PostgreSQL**: Direct interaction with the PostgreSQL database using GORM
- **Asset Models**: Uses Cloudinary asset models for data representation
- **Context**: Uses context for request lifecycle management
- **Transactions**: Supports database transactions for consistent operations
- **GORM**: Leverages GORM's features for ORM functionality, including soft deletes

## Data Flow

1. A request to perform an asset operation is received by the repository method
2. The method prepares the appropriate GORM query with parameters
3. The query is executed against the PostgreSQL database
4. Results are processed and returned to the caller
5. Any database errors are returned as appropriate

## Design Decisions

- **GORM ORM**: Uses GORM as the ORM layer for PostgreSQL operations
- **Soft Delete Pattern**: Implements soft deletes using GORM's deleted_at field
- **Unscoped Queries**: Uses Unscoped() for operations that need to access soft-deleted records
- **Pagination**: Supports limit/offset pagination for efficient data retrieval
- **Transaction Support**: Provides WithTx method for transaction management
- **Field Selection**: Supports partial field selection for optimized queries

See [API Reference](./api.md) for method-specific details.