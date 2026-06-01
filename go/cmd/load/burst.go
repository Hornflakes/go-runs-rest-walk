package main

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultBurstShardsCount = 40
	defaultBurstIntervalMs  = 5
)

type wsWriter struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func (w *wsWriter) writeText(data []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func (w *wsWriter) writeClose(code int, text string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	closeMsg := websocket.FormatCloseMessage(code, text)
	deadline := time.Now().Add(time.Second)
	return w.conn.WriteControl(websocket.CloseMessage, closeMsg, deadline)
}

type burstShard struct {
	conns map[int]*wsWriter
	mutex sync.Mutex
}

type burstPool struct {
	shards     []burstShard
	intervalMs time.Duration
	shootMsg   []byte
}

func newBurstPool(shardsCount, intervalMs int, shootMsg []byte) *burstPool {
	if shardsCount < 1 {
		shardsCount = defaultBurstShardsCount
	}
	if intervalMs <= 0 {
		intervalMs = defaultBurstIntervalMs
	}

	shards := make([]burstShard, shardsCount)
	for i := range shards {
		shards[i].conns = make(map[int]*wsWriter)
	}

	return &burstPool{shards: shards, intervalMs: time.Duration(intervalMs) * time.Millisecond, shootMsg: shootMsg}
}

func (pool *burstPool) register(clientId int, writer *wsWriter) {
	shard := &pool.shards[clientId%len(pool.shards)]
	shard.mutex.Lock()
	shard.conns[clientId] = writer
	shard.mutex.Unlock()
}

func (pool *burstPool) unregister(clientId int) {
	shard := &pool.shards[clientId%len(pool.shards)]

	shard.mutex.Lock()
	delete(shard.conns, clientId)
	shard.mutex.Unlock()
}

func (pool *burstPool) fireLoop(ctx context.Context) {
	var idx int

	for {
		if ctx.Err() != nil {
			return
		}

		tickStartTime := time.Now()
		shard := &pool.shards[idx%len(pool.shards)]
		idx++

		shard.mutex.Lock()
		for _, w := range shard.conns {
			_ = w.writeText(pool.shootMsg)
		}
		shard.mutex.Unlock()

		if diff := pool.intervalMs - time.Since(tickStartTime); diff > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(diff):
			}
		}
	}
}
