# CreateSignedUploadURL Flowchart

This document illustrates the low-level logic of the `UpdateOwners` method in the Cloudinary service.
See [api documentation](../api.md) for more details.

## Flowchart

```mermaid
graph TD
    A["Start"] --> B["Validate Request"]
    B --> C{"Request Valid?"}
    C --> |No| D["Return ErrInvalidArgument"]
    C --> |Yes| E["Create New Asset"]
    E --> F["Create asset in the DB"]
    F --> G{"Create successful?"}
    G --> |No| H["Return Error"]
    G --> |Yes| I["Save Asset Metadata to ArangoDB"]
    I --> J{"Save successful?"}
    J -->|No| K["Return Error"]
    J --> |Yes| L["Group Owners from asset Metadata by OwnerType"]
    L --> M["gRPC call to Image Service to notify about owner relations"]
    M --> N{"gRPC call successful?"}
    N --> |No| P["Return Error"]
    N --> |Yes| O["End"]
    D --> O
    H --> O
    P --> O
    K --> O
```
