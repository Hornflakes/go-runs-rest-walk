package com.hornflakes.gorunsrestwalk.server;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import io.netty.channel.Channel;
import io.netty.handler.codec.http.websocketx.TextWebSocketFrame;

import java.net.InetSocketAddress;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;

public class Socket {

    private final Channel channel;
    private long playerId;
    private final LinkedBlockingQueue<Message> in = new LinkedBlockingQueue<>();
    private final AtomicBoolean disconnected = new AtomicBoolean(false);
    private volatile boolean closed;

    Socket(Channel channel) {
        this.channel = channel;
    }

    public long playerId() { return playerId; }

    public void setPlayerId(long id) { this.playerId = id; }

    public String remoteAddr() {
        var addr = (InetSocketAddress) channel.remoteAddress();
        return addr.getAddress().getHostAddress() + ":" + addr.getPort();
    }

    public LinkedBlockingQueue<Message> in() { return in; }

    public boolean disconnected() { return disconnected.get(); }

    public boolean closed() { return closed; }

    public void send(Message msg) {
        String json;
        try {
            json = msg.marshal();
        } catch (JsonProcessingException e) {
            Logger.GLOBAL.softError("websocket message marshal failed", logDetail(e));
            return;
        }

        channel.writeAndFlush(new TextWebSocketFrame(json)).addListener(f -> {
            if (!f.isSuccess() && !disconnected.get()) {
                Logger.GLOBAL.hardError("websocket message write failed", logDetail(f.cause()));
            }
        });
    }

    public void close() {
        if (closed) return;
        closed = true;
        channel.close();
    }

    void pushMessage(Message msg) {
        in.offer(msg);
    }

    void markDisconnected() {
        if (disconnected.getAndSet(true)) return;
        channel.close();
    }

    String logDetail(Throwable err) {
        return Logger.playerWithAddr(playerId, remoteAddr()) + " err=" + err.getMessage();
    }
}
