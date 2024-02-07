# :game_die: A server for multiplayer dice rolling :game_die:
## Written in Go using Server Side Events

### Endpoints:
- `/listen?username="{name}"`
Subscribe to SSE. Returns an ID and a username as `"{ 'id': uint32, 'username': string }"`

    | Key   | Value  |
    |--------------- |
    | id   | uint32  |
    | username   | string  |

- `/register`
Create user and obtain JWT with POST

    POST body { " username " : " example " }
    
    Response: {  "ID": id, "token": jwt, "username": "username"}
    
    

- `/roll` 
Triggers a roll request with `"{ 'id': uint32, 'dice': uint8 }"`

    | Key  | Value    |
    |--------------- | --------------- |
    | id   | uint32   |
    | dice   | uint8   |

Responds with an SSE Event with ID of the roller and the result of the roll

