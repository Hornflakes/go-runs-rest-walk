package gameloop

import (
	"sync"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

type QueueMessage struct {
	From    uint
	Message server.Message
}

type GameQueue struct {
	messages []*QueueMessage
	killChan chan struct{}
	mutex    sync.Mutex
}

func NewGameQueue() *GameQueue {
	return &GameQueue{
		messages: make([]*QueueMessage, 0),
		killChan: make(chan struct{}),
		mutex:    sync.Mutex{},
	}
}

func (q *GameQueue) Start(s1, s2 server.Socket) {
	for {
		select {
		case msg := <-s1.In():
			q.mutex.Lock()
			q.messages = append(q.messages, &QueueMessage{
				1,
				msg.Message,
			})
			q.mutex.Unlock()
		case msg := <-s2.In():
			q.mutex.Lock()
			q.messages = append(q.messages, &QueueMessage{
				2,
				msg.Message,
			})
			q.mutex.Unlock()
		case <-q.killChan:
			return
		}
	}
}

func (q *GameQueue) Stop() {
	q.killChan <- struct{}{}
}

func (q *GameQueue) empty() bool {
	ret := true
	for _, msg := range q.messages {
		ret = ret && msg == nil
		if !ret {
			break
		}
	}
	return ret
}

func (q *GameQueue) Flush() []*QueueMessage {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.empty() {
		return nil
	}

	messages := q.messages
	q.messages = make([]*QueueMessage, 0)
	return messages
}
