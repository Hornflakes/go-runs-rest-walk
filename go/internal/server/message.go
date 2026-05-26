package server

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

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

type SocketMessage struct {
	Type    int
	Message Message
}

func fromSocket(msg []byte) (SocketMessage, error) {
	message, err := UnmarshalMessage(msg)
	if err != nil {
		return SocketMessage{}, err
	}
	return SocketMessage{
		Type:    websocket.TextMessage,
		Message: message,
	}, nil
}

func CreateSocketMessage(msgType int) SocketMessage {
	return SocketMessage{
		Type: websocket.TextMessage,
		Message: Message{
			Type: msgType,
			Msg:  "",
		},
	}
}
