package transport

import(
  "time"
  "github.com/gorilla/websocket"
  log "github.com/Sirupsen/logrus"
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

func (t *Transport) PingExpired() bool{
  pingExpires := t.LastPing.Add(t.PingInterval).Add(time.Second*5)
  return time.Now().After(pingExpires) || (!t.hasHelo && time.Now().After(t.ConnectedAt.Add(time.Second*5)))
}

func (t *Transport) ClientFields() log.Fields{
  f := log.Fields{
    "address": t.Connection.RemoteAddr(),
    "pingInterval": t.PingInterval,
    "lastPing": t.LastPing,
    "connectionId": t.ConnectionId,
    "connectedAt": t.ConnectedAt,
  }

  return f
}
