package plugin

import(
  "os"
)

func NewPlugin(name string) *Plugin{
  p := Plugin{}
  p.Name = name
  p.Instance, _ = os.Hostname()

  return &p
}
