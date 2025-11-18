# ArangoDB Cloudinary Metadata Repository API

This document describes the API provided by the ArangoDB Cloudinary Metadata repository, which handles Cloudinary asset metadata operations in ArangoDB. The repository defines core data access functionality for storing and retrieving asset metadata in ArangoDB.

For an overview of the repository, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../../architecture/).

## Overview

The ArangoDB Cloudinary Metadata repository is defined by the `Repository` interface and implemented by the `arangoRepository` struct. It provides methods for interacting with Cloudinary asset metadata stored in ArangoDB, including operations for creating, reading, updating, and deleting metadata records.

Key methods include:

- `EnsureCollection`: Creates the collection if it doesn't exist.
- `Get`: Retrieves the metadata for a specific asset.
- `ListUnownedIDs`: Retrieves the keys of all assets that have no owners.
- `ListByKeys`: Retrieves metadata for a list of asset keys.
- `CreateOwners`: Creates an asset's metadata with a new list of owners.
- `UpdateOwners`: Creates or updates an asset's metadata with a new list of owners.
- `DeleteOwners`: Deletes an asset's metadata.
- `CountUnowned`: Counts all assets that have no owners.

## EnsureCollection

The `EnsureCollection` method creates the cloudinary_asset_metadata collection if it doesn't exist.

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

## CreateOwners

The `CreateOwners` method creates an asset's metadata with a new list of owners.

### Input parameters

| Parameter | Type                                | Required | Description                           |
|-----------|-------------------------------------|----------|---------------------------------------|
| ctx       | context.Context                     | Required | Context for managing request lifecycle |
| key       | string                              | Required | Asset key for the metadata            |
| owners    | []metadatamodel.Owner               | Required | List of owners for the asset          |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if operation failed, nil otherwise |

## UpdateOwners

The `UpdateOwners` method creates or updates an asset's metadata with a new list of owners.

### Input parameters

| Parameter | Type                                | Required | Description                           |
|-----------|-------------------------------------|----------|---------------------------------------|
| ctx       | context.Context                     | Required | Context for managing request lifecycle |
| key       | string                              | Required | Asset key for the metadata            |
| owners    | []metadatamodel.Owner               | Required | List of owners for the asset          |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if operation failed, nil otherwise |

## DeleteOwners

The `DeleteOwners` method deletes an asset's metadata.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| key       | string          | Required | Asset key for the metadata to delete  |

### Output

| Type  | Description                     |
|-------|---------------------------------|
| error | Error if operation failed, nil otherwise |

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