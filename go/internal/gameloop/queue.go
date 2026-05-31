package gameloop

import (
	"sync"

	"github.com/hornflakes/go-runs-rest-walk/internal/server"
)

type QueueMessage struct {
	From    uint
	Message server.Message
}

type Queue struct {
	messages []*QueueMessage
	ack      chan struct{}
	stop     chan struct{}
	mutex    sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		messages: make([]*QueueMessage, 0),
		ack:      make(chan struct{}, 1),
		stop:     make(chan struct{}, 1),
		mutex:    sync.Mutex{},
	}
}

func (q *Queue) Start(s0, s1 server.Socket) {
	for {
		select {
		case msg := <-s0.In():
			q.mutex.Lock()
			q.messages = append(q.messages, &QueueMessage{
				1,
				msg.Message,
			})
			q.mutex.Unlock()

			select {
			case q.ack <- struct{}{}:
			default:
			}

		case msg := <-s1.In():
			q.mutex.Lock()
			q.messages = append(q.messages, &QueueMessage{
				2,
				msg.Message,
			})
			q.mutex.Unlock()

			select {
			case q.ack <- struct{}{}:
			default:
			}

		case <-q.stop:
			return
		}
	}
}

func (q *Queue) Stop() {
	q.stop <- struct{}{}
}

func (q *Queue) empty() bool {
	ret := true
	for _, msg := range q.messages {
		ret = ret && msg == nil
		if !ret {
			break
		}
	}
	return ret
}

func (q *Queue) Flush() []*QueueMessage {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.empty() {
		return nil
	}

	messages := q.messages
	q.messages = make([]*QueueMessage, 0)
	return messages
}
