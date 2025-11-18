# ArangoDB Mux Metadata Repository API

This document describes the API provided by the ArangoDB Mux Metadata repository, which handles Mux asset metadata operations in ArangoDB. The repository defines core data access functionality for storing and retrieving asset metadata in ArangoDB.

For an overview of the repository, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../../architecture/).

## Overview

The ArangoDB Mux Metadata repository is defined by the `Repository` interface and implemented by the `arangoRepository` struct. It provides methods for interacting with Mux asset metadata stored in ArangoDB, including operations for creating, reading, updating, and deleting metadata records.

Key methods include:

- `EnsureCollection`: Creates the collection if it doesn't exist.
- `Get`: Retrieves the metadata for a specific asset.
- `Create`: Creates an asset's metadata.
- `Update`: Creates or updates an asset's metadata with new values.
- `Delete`: Deletes an asset's metadata.
- `ListUnownedIDs`: Retrieves the keys of all assets that have no owners.
- `ListByKeys`: Retrieves metadata for a list of asset keys.
- `CountUnowned`: Counts all assets that have no owners.

## EnsureCollection

The `EnsureCollection` method creates the mux_asset_metadata collection if it doesn't exist.

### Input parameters

| Parameter | Type                | Required | Description                           |
|-----------|---------------------|----------|---------------------------------------|
| ctx       | context.Context     | Required | Context for managing request lifecycle |
| db        | arangodb.Database   | Required | Database instance to create collection in |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if creation failed, nil otherwise |

## Get

The `Get` method retrieves the metadata for a specific asset.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| key       | string          | Required | The key of the asset metadata to get    |

### Output

| Type                            | Description                              |
|---------------------------------|------------------------------------------|
| *metadatamodel.AssetMetadata   | Asset metadata if found, nil otherwise    |
| error                           | Error if operation failed, nil otherwise |

## Create

The `Create` method creates an asset's metadata.

### Input parameters

| Parameter | Type                           | Required | Description                           |
|-----------|--------------------------------|----------|---------------------------------------|
| ctx       | context.Context                | Required | Context for managing request lifecycle |
| metadata  | *metadatamodel.AssetMetadata  | Required | The metadata to create                |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if operation failed, nil otherwise |

## Update

The `Update` method creates or updates an asset's metadata with new values. It performs a partial update, only modifying fields that are not nil in the provided metadata struct.

### Input parameters

| Parameter | Type                           | Required | Description                           |
|-----------|--------------------------------|----------|---------------------------------------|
| ctx       | context.Context                | Required | Context for managing request lifecycle |
| key       | string                         | Required | Asset key for the metadata to update  |
| metadata  | *metadatamodel.AssetMetadata  | Required | The metadata with updated values      |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if operation failed, nil otherwise |

## Delete

The `Delete` method deletes an asset's metadata.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| key       | string          | Required | Asset key for the metadata to delete  |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if operation failed, nil otherwise |

## ListUnownedIDs

The `ListUnownedIDs` method retrieves the keys of all assets that have no owners.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |

### Output

| Type         | Description                              |
|--------------|------------------------------------------|
| []string     | List of asset keys with no owners        |
| error        | Error if operation failed, nil otherwise |

## ListByKeys

The `ListByKeys` method retrieves metadata for a list of asset keys.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| keys      | []string        | Required | List of asset keys to retrieve        |

### Output

| Type                                          | Description                              |
|-----------------------------------------------|------------------------------------------|
| map[string]*metadatamodel.AssetMetadata      | Map of asset keys to metadata            |
| error                                         | Error if operation failed, nil otherwise |

## CountUnowned

The `CountUnowned` method counts all assets that have no owners.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of unowned assets                 |
| error  | Error if operation failed, nil otherwise |