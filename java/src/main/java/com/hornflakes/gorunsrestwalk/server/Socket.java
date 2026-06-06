package com.hornflakes.gorunsrestwalk.server;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import org.eclipse.jetty.websocket.api.Callback;
import org.eclipse.jetty.websocket.api.Session;
import org.eclipse.jetty.websocket.api.annotations.*;

import java.time.Duration;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

@WebSocket
public class Socket {

    private static final Message OUT_CLOSE_SENTINEL = new Message(Message.CLOSE_OUT);
    private static final Message IN_CLOSE_SENTINEL = new Message(Message.CLOSE_IN);
    private static final long WRITE_DRAIN_TIMEOUT_MS = 5_000;

    private Session session;
    private String remoteAddr;
    private volatile long playerId;
    private final BlockingQueue<Message> in = new LinkedBlockingQueue<>();
    private final BlockingQueue<Message> out = new LinkedBlockingQueue<>();
    private final AtomicBoolean disconnected = new AtomicBoolean(false);
    private final AtomicBoolean closeOnce = new AtomicBoolean(false);
    private volatile boolean closed;
    private Thread writeThread;
    private Runnable onReady;

    public static boolean isInboundClosed(Message msg) {
        return msg == IN_CLOSE_SENTINEL;
    }

    public long playerId() { return playerId; }

    public void setPlayerId(long id) { this.playerId = id; }

    public String remoteAddr() { return remoteAddr; }

    public boolean disconnected() { return disconnected.get(); }

    public boolean closed() { return closed; }

    public Message pollInbound(long timeout, TimeUnit unit) throws InterruptedException {
        return in.poll(timeout, unit);
    }

    public void setOnReady(Runnable onReady) { this.onReady = onReady; }

    public void send(Message msg) {
        if (disconnected.get() || closed) return;
        try {
            out.put(msg);
        } catch (InterruptedException _) {
            Thread.currentThread().interrupt();
        }
    }

    public void close() {
        if (!closeOnce.compareAndSet(false, true)) return;
        closed = true;

        try {
            out.put(OUT_CLOSE_SENTINEL);
        } catch (InterruptedException _) {
            Thread.currentThread().interrupt();
        }

        if (writeThread != null) {
            try {
                writeThread.join(WRITE_DRAIN_TIMEOUT_MS);
            } catch (InterruptedException _) {
                Thread.currentThread().interrupt();
            }
        }

        closeSession();
    }


    @OnWebSocketOpen
    public void onOpen(Session session) {
        this.session = session;
        session.setIdleTimeout(Duration.ZERO);
        this.remoteAddr = session.getRemoteSocketAddress().toString();
        if (remoteAddr.startsWith("/")) {
            remoteAddr = remoteAddr.substring(1);
        }

        writeThread = Thread.ofVirtual().start(() -> {
            try {
                while (true) {
                    Message msg = out.take();
                    if (msg == OUT_CLOSE_SENTINEL) break;

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
                Thread.currentThread().interrupt();
            }
        });

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
            Thread.currentThread().interrupt();
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


    private void markDisconnected() {
        if (disconnected.getAndSet(true)) return;
        closed = true;
        in.offer(IN_CLOSE_SENTINEL);
        closeSession();
    }

    private void closeSession() {
        try {
            if (session != null && session.isOpen()) {
                session.close();
            }
        } catch (Exception _) {
        }
    }

    private static boolean normalClose(Throwable err) {
        if (err == null) return true;
        String msg = err.getMessage();
        if (msg == null) return false;
        return msg.contains("NORMAL")
            || msg.contains("GOING_AWAY")
            || msg.contains("closed")
            || msg.contains("Idle Timeout");
    }

    private String logDetail(Throwable err) {
        return Logger.playerWithAddr(playerId, remoteAddr) + " err=" + err.getMessage();
    }
}
