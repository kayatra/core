package transport

import(
  "time"
  "github.com/gorilla/websocket"
)

type Transport struct{
  Connection      *websocket.Conn
  ConnectionId    uint64
  LastPing        time.Time
  PingInterval    time.Duration
  ConnectedAt     time.Time

  wsUrl           string
  hasHelo         bool
  commandChannel  chan Command
  connectionMade  bool
  disconnectError error
}
