package com.hornflakes.gorunsrestwalk.server;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import org.eclipse.jetty.websocket.api.Callback;
import org.eclipse.jetty.websocket.api.Session;
import org.eclipse.jetty.websocket.api.annotations.*;

import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;

@WebSocket
public class Socket {

    private Session session;
    private String remoteAddr;
    private volatile long playerId;
    private final BlockingQueue<Message> in = new ArrayBlockingQueue<>(1);
    private final BlockingQueue<Message> out = new ArrayBlockingQueue<>(1);
    private final AtomicBoolean disconnected = new AtomicBoolean(false);
    private volatile boolean closed;
    private Thread writeThread;
    private Runnable onReady;

    public long playerId() { return playerId; }
    public void setPlayerId(long id) { this.playerId = id; }
    public String remoteAddr() { return remoteAddr; }
    public boolean disconnected() { return disconnected.get(); }
    public boolean closed() { return closed; }
    public BlockingQueue<Message> in() { return in; }
    public BlockingQueue<Message> out() { return out; }
    public void setOnReady(Runnable onReady) { this.onReady = onReady; }

    @OnWebSocketOpen
    public void onOpen(Session session) {
        this.session = session;
        this.remoteAddr = session.getRemoteSocketAddress().toString();
        if (remoteAddr.startsWith("/")) {
            remoteAddr = remoteAddr.substring(1);
        }

        writeThread = Thread.ofVirtual().start(() -> {
            try {
                while (true) {
                    Message msg = out.take();
                    String json;
                    try {
                        json = msg.marshal();
                    } catch (Exception e) {
                        Log.softError("websocket message marshal failed", logDetail(e));
                        continue;
                    }

                    try {
                        session.sendText(json, Callback.NOOP);
                    } catch (Exception e) {
                        if (!disconnected.get()) {
                            Log.hardError("websocket message write failed", logDetail(e));
                        }
                        break;
                    }
                }
            } catch (InterruptedException _) {
            }
        });

        send(Message.createHello(playerId));
        Log.info("websocket connected", Logger.playerWithAddr(playerId, remoteAddr));

        if (onReady != null) {
            onReady.run();
        }
    }

    @OnWebSocketMessage
    public void onMessage(String text) {
        Message msg;
        try {
            msg = Message.unmarshal(text);
        } catch (Exception e) {
            Log.softError("websocket message unmarshal failed", logDetail(e));
            return;
        }

        try {
            in.put(msg);
        } catch (InterruptedException _) {
        }
    }

    @OnWebSocketClose
    public void onClose(int statusCode, String reason) {
        markDisconnected();
    }

    @OnWebSocketError
    public void onError(Throwable cause) {
        if (!normalClose(cause)) {
            Log.warn("websocket read ended", logDetail(cause));
        }
        markDisconnected();
    }

    public void close() {
        if (closed) return;
        closed = true;

        if (writeThread != null) {
            writeThread.interrupt();
        }

        try {
            session.close();
        } catch (Exception _) {
        }
    }

    public void send(Message msg) {
        if (disconnected.get() || closed) return;
        try {
            out.put(msg);
        } catch (InterruptedException _) {
        }
    }

    private void markDisconnected() {
        if (disconnected.getAndSet(true)) return;
        try {
            session.close();
        } catch (Exception _) {
        }
    }

    private static boolean normalClose(Throwable err) {
        if (err == null) return true;
        String msg = err.getMessage();
        if (msg == null) return false;
        return msg.contains("NORMAL") || msg.contains("GOING_AWAY") || msg.contains("closed");
    }

    private String logDetail(Throwable err) {
        return Logger.playerWithAddr(playerId, remoteAddr) + " err=" + err.getMessage();
    }
}
