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

  t.lastPing = time.Now()
  t.pingInterval = time.Minute
  t.hasHelo = false
  t.connectedAt = time.Now()

  c, _, err := websocket.DefaultDialer.Dial(t.wsUrl, nil)
  t.connection = c
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
    pingExpires := t.lastPing.Add(t.pingInterval).Add(time.Second*5)
    log.WithFields(log.Fields{
      "pingInterval": t.pingInterval,
      "lastPing": t.lastPing,
      "pingExpires": pingExpires,
      "connectionFailures": connectionFailures,
      "connectionId": t.connectionId,
      "connectedAt": t.connectedAt,
    }).Debug("Checking connection")

    if time.Now().After(pingExpires) || (!t.hasHelo && time.Now().After(t.connectedAt.Add(time.Second*5))){
      if t.connectionMade{
        t.connection.Close()
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
    msgType, msg, err := t.connection.ReadMessage()
    if err != nil{
      if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
        log.WithFields(log.Fields{
          "id": t.connectionId,
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
      cmdProcessors := map[string]func(*Command){
        "helo": t.msgHelo,
      }

      if cmdProcessors[payload.Type] == nil{
        log.WithFields(log.Fields{
          "command": payload.Type,
          "payload": string(msg),
        }).Error("Could not find processor for command")
      } else {
        cmdProcessors[payload.Type](&payload)
      }
    } else {
      log.WithFields(log.Fields{
        "msg": string(msg),
        "err": err,
      }).Warning("Could not read command from controller")
    }
  }
}
