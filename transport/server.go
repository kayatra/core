package transport

import(
  "github.com/gorilla/websocket"
  "time"
  log "github.com/Sirupsen/logrus"
  "net/http"
  "sync/atomic"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
}

const pingInterval = time.Second*5
var connectionIdCounter uint64 = 0
var commandIdCounter uint64 = 0

type socketConnection struct{
  *Transport
}

type socketMsg struct{
  Type              string      `json:"message_type"`
  Body              interface{} `json:"body"`
  CommandId         uint64      `json:"command_id"`
  Time              time.Time   `json:"time"`
}

func (c *socketConnection) NewMessage(mt string, b interface{}) *socketMsg{
  m := socketMsg{}
  m.CommandId = atomic.AddUint64(&commandIdCounter, 1)
  m.Type = mt
  m.Time = time.Now().UTC()
  m.Body = b

  return &m
}

func (c *socketConnection) SendMessage(mt string, b interface{}){
  msg := c.NewMessage(mt, b)

  log.WithFields(log.Fields{
    "id": c.ConnectionId,
    "type": mt,
    "body": b,
  }).Debug("Sending message")

  err := c.Connection.WriteJSON(msg)
  if err != nil {
    f := c.ClientFields()
    f["error"] = err
    f["type"] = mt
    f["body"] = b
    log.WithFields(f).Error("Error sending message to client")
  }
}

func ping(c socketConnection, close chan struct{}) {
  ticker := time.NewTicker(pingInterval)
  defer ticker.Stop()
  for {
    select{
      case <-ticker.C:
        p := CommandPing{}
        c.SendMessage("ping", p)

        if c.PingExpired(){
          log.WithFields(c.ClientFields()).Warning("Ping expired for client")
        }
      case <-close:
        return
    }
  }
}

func Server(w http.ResponseWriter, r *http.Request) {
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.WithFields(log.Fields{
      "err": err,
    }).Error("Failed to set websocket upgrade")
    return
  }

  defer ws.Close()

  c := socketConnection{}
  t := Transport{}
  t.Connection = ws
  t.ConnectionId = atomic.AddUint64(&connectionIdCounter, 1)
  c.Transport = &t

  log.WithFields(t.ClientFields()).Debug("New socket connection")

  done := make(chan struct{})

  h := CommandHelo{}
  h.ConnectionId = c.ConnectionId
  h.PingInterval = pingInterval.Seconds()
  c.SendMessage("helo", h)

  go ping(c, done)

  for{
    msgType, msg, err := ws.ReadMessage()
    if err != nil{
      if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
        log.WithFields(c.ClientFields()).Debug("Client disconnect")
      } else {
        f := c.ClientFields()
        f["err"] = err
        log.WithFields(f).Error("Error reading message")
      }
      break
    }

    f := c.ClientFields()
    f["msgType"] = msgType
    f["msg"] = msg
    log.WithFields(f).Debug("Got message from client")
  }

  log.WithFields(log.Fields{
    "id": c.ConnectionId,
  }).Debug("Closing client connection")

  close(done)
}
