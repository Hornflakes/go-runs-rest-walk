package server

import "encoding/json"

const (
	Ready int = iota
	GameOn
	Shoot
	GameOver
)

type Message struct {
	Type int    `json:"type"`
	Msg  string `json:"msg,omitempty"`
}

func UnmarshalMessage(msg []byte) (Message, error) {
	var message Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		return Message{}, err
	}
	return message, nil
}

func CreateMessage(msgType int) Message {
	return Message{
		Type: msgType,
		Msg:  "",
	}
}
