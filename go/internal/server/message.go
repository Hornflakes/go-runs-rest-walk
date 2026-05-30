package server

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/hornflakes/go-runs-rest-walk/internal/stats"
)

const (
	Hello int = iota
	Ready
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

func CreateHelloMessage(playerId uint64) SocketMessage {
	return SocketMessage{
		Type: websocket.TextMessage,
		Message: Message{
			Type: Hello,
			Msg:  fmt.Sprintf("%d", playerId),
		},
	}
}

func CreateReadyMessage(enemyId uint64) SocketMessage {
	return SocketMessage{
		Type: websocket.TextMessage,
		Message: Message{
			Type: Ready,
			Msg:  fmt.Sprintf("enemyId=%d", enemyId),
		},
	}
}

func CreateWinnerMessage(gameFrameStats *stats.GameFrameStats) SocketMessage {
	return SocketMessage{
		Type: websocket.TextMessage,
		Message: Message{
			Type: GameOver,
			Msg:  fmt.Sprintf("winner(%d)->%s", stats.ActiveGames, gameFrameStats),
		},
	}
}

func CreateLoserMessage() SocketMessage {
	return SocketMessage{
		Type: websocket.TextMessage,
		Message: Message{
			Type: GameOver,
			Msg:  "loser",
		},
	}
}
