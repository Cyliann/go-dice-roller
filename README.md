# :game_die: A backend for multiplayer dice rolling :game_die:

## Written in Go using Server-Sent Events

### Endpoints:

- `/play`
  Subscribe to SSE. Returns an ID of a subscribed stream `'{ "room": string }'`

| Key  | Value  |
| ---- | ------ |
| room | string |

- `/roll`
  Triggers a roll request with `'{ "dice": { "id": nrOfSides } }'` (allows for multiple rolls and keeps the order) 

| Key  | Value                |
| ---- |----------------------|
| dice | { "id" : nrOfSides } |

Example body of request:
`'{ "dice": { "0": 6, "1": 10, "2": 20 } }'`

Responds with an SSE Event with a username of the roller, id of the stream and the result of the roll `'{ "username": string, "room": string, "result": uint8 }'`

| Key      | Value  |
| -------- | ------ |
| username | string |
| room     | string |
| result   | int    |
