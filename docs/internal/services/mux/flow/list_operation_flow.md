# List operation Flowchart

This document illustrates the low-level logic of the all `List` methods in the Mux service:

- `List`
- `ListDeleted`
- `ListUnowned`

See [api documentation](../api.md) for more details.

## Flowchart

```mermaid
graph TD
    A((Start)) --> B[Retrieve assets from the Database]
    B --> C{Database operation Successful?}
    C --> |No| D[Return Error]
    C --> |Yes| E[Count assets in the Database]
    E --> F{Database operation Successful?}
    F --> |No| G[Return Error]
    F --> |Yes| H[Return assets and count]
    H --> I[End]
    D --> I
    G --> I
```
