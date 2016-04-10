package transport

import(
  log "github.com/Sirupsen/logrus"
  "encoding/json"
  "time"
)

type Command struct{
  Id          int64         `json:"command_id"`
  Type        string        `json:"message_type"`
  Body        interface{}   `json:"body"`
  Transport   *Transport
}

type commandInterface interface{
  SetCommand(*Command)
  Process()
}

type baseCommand struct{
  *Command
}

func (c *baseCommand) SetCommand(cmd *Command){
  c.Command = cmd
}

func processCommand(c *Command){
  // I'm sure there's a more efficient way of doing this
  var m commandInterface
  raw, _ := json.Marshal(c.Body)
  switch c.Type{
  case "helo":
    m = &CommandHelo{}
  }

  err := json.Unmarshal(raw, &m)
  if err != nil{
    log.WithFields(log.Fields{
      "err": err,
      "command": c.Type,
      "body": c.Body,
    }).Error("Could not process command")
  } else {
    m.SetCommand(c)
    m.Process()
  }
}

type CommandHelo struct{
  baseCommand
  PingInterval    float64   `json:"ping_interval"`
  ConnectionId    uint64    `json:"connection_id"`
}

func (c CommandHelo) Process(){
  c.Command.Transport.hasHelo = true
  c.Command.Transport.connectionId = c.ConnectionId
  c.Command.Transport.pingInterval = time.Duration(c.PingInterval)*time.Second
  log.Debug(c.PingInterval)

  log.WithFields(log.Fields{
    "connectionId": c.Command.Transport.connectionId,
    "pingInterval": c.Command.Transport.pingInterval,
  }).Debug("Got helo from controller")
}
