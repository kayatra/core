package transport

import(
  log "github.com/Sirupsen/logrus"
  "time"
)

type HeloCommand struct{
  PingInterval    time.Duration   `json:"ping_interval"`
  ConnectionId    uint64          `json:"connection_id"`
}

func (t *Transport) msgHelo(p *Command){
  t.hasHelo = true
  log.Debug(p.Body)
  heloData := p.Body.(HeloCommand)
  t.connectionId = heloData.ConnectionId
  t.pingInterval = heloData.PingInterval

  log.WithFields(log.Fields{
    "connectionId": t.connectionId,
    "pingInterval": t.pingInterval,
  }).Debug("Got helo from controller")
}
