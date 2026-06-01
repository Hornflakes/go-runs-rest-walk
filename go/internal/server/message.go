package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
			Msg:  fmt.Sprintf("playerId=%d", playerId),
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

func CreateWinnerMessage(winnerId uint64, gameFrameStats *stats.GameFrameStats) SocketMessage {
	return SocketMessage{
		Type: websocket.TextMessage,
		Message: Message{
			Type: GameOver,
			Msg: fmt.Sprintf(
				"winner=%d histogram=%s active_games=%d",
				winnerId,
				gameFrameStats,
				stats.ActiveGames,
			),
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

func parsePrefixedId(msg, prefix string) (uint64, error) {
	rest := strings.TrimPrefix(msg, prefix)
	return strconv.ParseUint(rest, 10, 64)
}

func ParseHelloMessage(msg string) (uint64, error) {
	return parsePrefixedId(msg, "playerId=")
}

func ParseReadyMessage(msg string) (uint64, error) {
	return parsePrefixedId(msg, "enemyId=")
}
