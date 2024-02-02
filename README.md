# 󰝮 A server for multiplayer dice rolling 󱅕
## Written in Go using Server Side Events

### Endpoints:
- `/listen`
Subscribe to SSE. Returns an ID as `"{ 'id': uint32 }"`

    | Key   | Value  |
    |--------------- | --------------- |
    | id   | uint32  |
- `/roll` 
Triggers a roll request with `"{ 'id': uint32, 'dice': uint8 }"`

    | Key  | Value    |
    |--------------- | --------------- |
    | id   | uint32   |
    | dice   | uint8   |

