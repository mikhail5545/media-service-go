# CreateSignedUploadURL Flowchart

This document illustrates the low-level logic of the `UpdateOwners` method in the Cloudinary service.
See [api documentation](../api.md) for more details.

## Flowchart

```mermaid
graph TD
    A["Start"] --> B["Validate Request"]
    B --> C{"Is Request Valid?"}
    C --> |No| D["Return ErrInvalidArgument"]
    C --> |Yes| E["Retrieve asset from the DB"]
    E --> F{"Asset Found?"}
    F --> |No| G["Return ErrNotFound"]
    F --> |Yes| H["Retrieve Asset Metadata From ArangoDB"]
    H --> I{"Successful DB call?"}
    I --> |No| J["Return Error"]
    I --> |Yes| K{"Does Asset Have Owners?"}
    K --> |No| L["Set currentOwners as blank slice"]
    K --> |Yes| M["Set currentOwnes from asset Metadata"]
    M --> N["Group Current Owners By OwnerType"]
    L --> N
    N --> O["Group New owners from request By OwnerType"]
    O --> P["Update Asset Metadata in the ArangoDB"]
    P --> Q{"Update successful?"}
    Q --> |No| R["Return Error"]
    Q --> |Yes| S["Call gRPC Image Service to notify about changes"]
    S --> T{"gRPC call Successful?"}
    T --> |No| Y["Return Error"]
    T --> |Yes| X["End"]
    Y --> X
    D --> X
    J --> X
    G --> X
    R --> X
```
