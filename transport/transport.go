package transport

import(
  "time"
  "github.com/gorilla/websocket"
)

type Transport struct{
  connection      *websocket.Conn
  wsUrl           string
  hasHelo         bool
  commandChannel  chan Command
  connectionId    uint64
  lastPing        time.Time
  pingInterval    time.Duration
  connectedAt     time.Time
  connectionMade  bool
  disconnectError error
}
