# ArangoDB Package API

This document describes the API provided by the ArangoDB package, which handles ArangoDB connections in the microservice. The package defines core functionality for establishing connections and creating databases in ArangoDB.

For an overview of the package, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../architecture/).

## Overview

The ArangoDB package provides functions to:

- `NewArangoDB`: Initialize a connection to an existing ArangoDB database named "media_service"
- `CreateArangoDB`: Create a new ArangoDB database instance with specified configuration

## NewArangoDB

The `NewArangoDB` function initializes and returns a connection to an existing ArangoDB database named "media_service".

### Input parameters

| Parameter | Type         | Required | Description                           |
|-----------|--------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| e         | []string     | Required | List of endpoints for ArangoDB instances |

### Output

| Type                | Description                                    |
|---------------------|------------------------------------------------|
| arangodb.Database   | Interface to the connected ArangoDB database   |
| error               | Error if connection failed, nil otherwise       |

### Example

```go
endpoints := []string{"http://localhost:8529"}
db, err := arango.NewArangoDB(context.Background(), endpoints)
if err != nil {
    // Handle error
}
// Use db for database operations
```

## CreateArangoDB

The `CreateArangoDB` function creates a new ArangoDB database instance with the specified name.

### Input parameters

| Parameter | Type                | Required | Description                                      |
|-----------|---------------------|----------|--------------------------------------------------|
| ctx       | context.Context     | Required | Context for managing request lifecycle           |
| name      | string              | Required | Name for the new database                        |
| c         | arangodb.Client     | Required | ArangoDB client to use for database creation     |

### Output

| Type                | Description                                    |
|---------------------|------------------------------------------------|
| arangodb.Database   | Interface to the created ArangoDB database     |
| error               | Error if creation failed, nil otherwise         |

### Example

```go
database, err := arango.CreateArangoDB(context.Background(), "media_service", client)
if err != nil {
    // Handle error
}
// Use database for operations
```