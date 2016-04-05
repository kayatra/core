package plugin

import(
  "os"
  log "github.com/Sirupsen/logrus"
  "github.com/voxelbrain/goptions"
  "github.com/kayatra/core/transport"
)

type options struct {
  Controller  string        `goptions:"-c, --controller, description='URL of the controller. Also loaded from the HC_CONTROLLER environment variable'"`
  Instance    string        `goptions:"-i, --instance, description='Instance identifier for plugin.'"`
  Verbose     bool          `goptions:"-v, --verbose, description='Log verbosely'"`
  Help        goptions.Help `goptions:"-h, --help, description='Show help'"`
}

func (p *Plugin) Start(){
  parsedOptions := options{}
  controllerEnv := os.Getenv("HC_CONTROLLER")
  if controllerEnv == ""{
    controllerEnv = "http://localhost:9211/"
  }
  parsedOptions.Controller = controllerEnv

  parsedOptions.Instance = p.Instance

  goptions.ParseAndFail(&parsedOptions)

  log.SetFormatter(&log.TextFormatter{FullTimestamp:true})

  if parsedOptions.Verbose{
    log.SetLevel(log.DebugLevel)
    log.Debug("Logging verbosely")
  } else {
    log.SetLevel(log.InfoLevel)
  }

  p.Instance = parsedOptions.Instance
  p.commandChannel = make(chan transport.Command)

  logFields := log.Fields{
    "name": p.Name,
    "instance": p.Instance,
    "controller": parsedOptions.Controller,
  }
  log.WithFields(logFields).Info("Starting plugin")

  t, err := transport.New(parsedOptions.Controller, p.commandChannel)

  if err != nil {
    logFields["err"] = err
    log.WithFields(logFields).Error("Could not connect to controller")
    os.Exit(1)
  }

  p.Transport = &t

  p.RecieveCommands()
}
