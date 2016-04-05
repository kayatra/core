package transport

type Command struct{
  Id      int64         `json:"command_id"`
  Type    string        `json:"message_type"`
  Body    interface{}   `json:"body"`
}
