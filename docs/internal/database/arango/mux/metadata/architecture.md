# ArangoDB Mux Metadata Repository Architecture

This document describes the technical design of the ArangoDB Mux Metadata repository, which manages Mux asset metadata within the microservice.

## Repository Architecture

The ArangoDB Mux Metadata repository is defined by the `Repository` interface and implemented by the `arangoRepository` struct. It consists of:

- **Data Access Layer**: Provides methods for interacting with the `mux_asset_metadata` collection in ArangoDB
- **Query Management**: Handles AQL queries for retrieving, updating, and managing asset metadata
- **Partial Updates**: Supports selective updates of metadata fields
- **Error Handling**: Manages ArangoDB-specific errors and converts them to domain errors

The repository follows a standard repository pattern architecture:

1. **Interface Definition**: Defines the contract for metadata operations
2. **Implementation**: Provides concrete ArangoDB-based implementation
3. **Query Execution**: Executes AQL queries against the database
4. **Result Handling**: Processes and returns results

## Interactions

- **ArangoDB**: Direct interaction with the ArangoDB database using AQL queries
- **Asset Models**: Uses Mux asset metadata models for data representation
- **Context**: Uses context for request lifecycle management
- **Error Handling**: Returns domain-specific errors like `ErrNotFound` and `ErrConflict`

## Data Flow

1. A request to perform a metadata operation is received by the repository method
2. The method prepares the appropriate AQL query with parameters
3. The query is executed against the ArangoDB instance
4. Results are processed and returned to the caller
5. Any database errors are wrapped and returned as appropriate

## Design Decisions

- **Collection Name**: Uses "mux_asset_metadata" as the fixed collection name
- **Key-Based Access**: Uses the asset key as the document key in ArangoDB
- **AQL Queries**: Uses AQL for efficient queries including filtering for unowned assets
- **Partial Updates**: Supports field-specific updates using dynamic query building
- **Error Mapping**: Maps ArangoDB-specific errors to domain errors for consistency

See [API Reference](./api.md) for method-specific details.