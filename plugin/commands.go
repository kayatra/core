package plugin

import(
  log "github.com/Sirupsen/logrus"
)

func (p *Plugin) RecieveCommands(){
  for{
    log.Debug(<- p.commandChannel)
  }
}
