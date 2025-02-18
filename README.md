# :game_die: A backend for multiplayer dice rolling :game_die:

## Written in Go using Server-Sent Events

### Endpoints:
- `/docs/index.html`
  [Swagger](https://github.com/swaggo/swag) documentation.

- `/register` POST
  Register a new user in a specific room with `'{ "user": string, "room": string }`

| Key  | Value    |
| ---- | -------- |
| user | string   |
| room | string   |

  If _room_ is an empty string it creates a new room.
  If _room_ is non empty and doesn't exist, returns a 400:"Room doesn't exist" error.

  Returns a room in which the user is registered.
  
| Key  | Value    |
| ---- | -------- |
| room | string   |

- `/play` GET
  Subscribe to SSE.

- `/roll` POST
  Triggers a roll request with `'{ "dice": uin8 }'` 

| Key  | Value   |
| ---- | ------- |
| dice | uin8    |

Example body of request:
`'{ "dice": 20 }'`

Responds with an SSE Event with a username of the roller, id of the stream and the result of the roll `'{ "username": string, "room": string, "result": uint8 }'`

| Key      | Value  |
| -------- | ------ |
| username | string |
| room     | string |
| result   | int    |
