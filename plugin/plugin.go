package plugin

import(
  "github.com/kayatra/core/transport"
)

type Plugin struct{
  Name              string
  Instance          string
  Transport         *transport.Transport
  commandChannel    chan transport.Command
}
