package transport

import(
  log "github.com/Sirupsen/logrus"
  "github.com/gorilla/websocket"
  "time"
  "encoding/json"
)

func (t *Transport) establish(){
  t.connect()
  go t.ensureConnection()
}

func (t *Transport) connect() error{
  log.WithFields(log.Fields{
    "wsUrl": t.wsUrl,
  }).Debug("Connecting to transport websocket")

  t.LastPing = time.Now()
  t.PingInterval = time.Minute
  t.ConnectedAt = time.Now()
  t.hasHelo = false

  c, _, err := websocket.DefaultDialer.Dial(t.wsUrl, nil)
  t.Connection = c
  if err != nil {
    log.WithFields(log.Fields{
      "err": err,
      "socketUrl": t.wsUrl,
    }).Error("Could not connect to controller")
    return err
  }

  t.connectionMade = true

  go t.readCommands()

  return nil
}

const reconnectPause = time.Second*5
const maxRecconnectPause = time.Minute*3
func (t *Transport) ensureConnection(){
  var connectionFailures int

  tkr := time.NewTicker(time.Second*5)
  for {
    <- tkr.C
    f := t.ClientFields()
    f["connectionFailures"] = connectionFailures
    log.WithFields(f).Debug("Checking connection")

    if t.PingExpired(){
      if t.connectionMade{
        t.Connection.Close()
        t.connectionMade = false
      }
      t.hasHelo = false

      reconnectIn := reconnectPause*time.Duration(connectionFailures)
      log.Debug(reconnectIn)
      if reconnectIn.Seconds() > maxRecconnectPause.Seconds(){
        reconnectIn = maxRecconnectPause
      }
      log.WithFields(log.Fields{
        "connectpause": reconnectIn,
      }).Warning("Lost connection to controller")
      time.Sleep(reconnectIn)
      t.connect()

      connectionFailures++
    } else {
      if t.hasHelo && connectionFailures > 0{
        connectionFailures = 0
      }
    }
  }
}


func (t *Transport) readCommands(){
  log.Debug("Reading commands")
  for{
    msgType, msg, err := t.Connection.ReadMessage()
    if err != nil{
      if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
        log.WithFields(log.Fields{
          "id": t.ConnectionId,
        }).Debug("Lost connection to server")
      } else {
        log.WithFields(log.Fields{
          "err": err,
        }).Warning("Error reading message")
      }
      t.disconnectError = err
      break
    }

    log.WithFields(log.Fields{
      "type": msgType,
      "msg": string(msg),
    }).Debug("Got message from controller")
    payload := Command{}
    err = json.Unmarshal(msg, &payload)
    if err == nil{
      payload.Transport = t
      processCommand(&payload)
    } else {
      log.WithFields(log.Fields{
        "msg": string(msg),
        "err": err,
      }).Warning("Could not read command from controller")
    }
  }
}
