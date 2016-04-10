package transport

import(
  "net/url"
)

func New(controller string, commandChannel chan Command) (Transport, error){
  t := Transport{}
  t.commandChannel = commandChannel

  wsUrl, err := url.Parse(controller)
  if err != nil {
    return t, err
  }

  if wsUrl.Scheme == "http"{
    wsUrl.Scheme = "ws"
  } else if wsUrl.Scheme == "https"{
    wsUrl.Scheme = "wss"
  }

  wsUrl.Path = "/plugin/transport"

  t.wsUrl = wsUrl.String()

  t.establish()
  return t, nil
}
