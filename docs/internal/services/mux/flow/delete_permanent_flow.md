# DeletePermanent Flowchart

This document illustrates the low-level logic of the `DeletePermanent` method in the Mux service.

See [api documentation](../api.md) for more details.

## Flowchart

```mermaid
graph TD
    A((Start)) --> B[Validate AssetID]
    B --> C{Is AssetID a valid UUID?}
    C --> |No| D[Return ErrInvalidArgument]
    C --> |Yes| E[Start DB Transaction]
    E --> F[Retrieve asset from the DB]
    F --> G{Asset Found?}
    G -->|No| H[Return ErrNotFound]
    G -->|Yes| I{Is Asset's MuxUploadID field populated?}
    I --> |Yes| J[Call Mux API]
    J --> K[Delete Mux Asset]
    K --> L{API Call Successful?}
    L --> |No| M[Return ErrMuxAPI]
    L --> |Yes| N[Delete Asset Metadata From the ArangoDB]
    N --> O{Metadata Deleted Successful?}
    O --> |No| P[Return Error]
    O --> |Yes| Q[Delete Asset From the DB]
    Q --> R{Asset Deleted Successfuly?}
    R --> |No| S[Return Error]
    R --> |Yes| T[Commit DB Transaction]
    T --> Y[End]
    D --> Y
    H --> Y
    M --> Y
    P --> Y
    S --> Y
    I --> |No| N
```
