package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;

public class Queue {

    private static final long INBOUND_POLL_MS = 50;

    private final List<QueueMessage> messages = new ArrayList<>();
    private final Object mutex = new Object();
    private volatile boolean stopped;
    private Thread readerThread;

    public record QueueMessage(int from, Message message) {}

    public void start(Socket s0, Socket s1) {
        readerThread = Thread.ofVirtual().start(() -> {
            while (!stopped) {
                try {
                    Message msg = s0.pollInbound(INBOUND_POLL_MS, TimeUnit.MILLISECONDS);
                    if (msg != null) {
                        if (Socket.isInboundClosed(msg)) return;
                        append(1, msg);
                        continue;
                    }

                    msg = s1.pollInbound(INBOUND_POLL_MS, TimeUnit.MILLISECONDS);
                    if (msg != null) {
                        if (Socket.isInboundClosed(msg)) return;
                        append(2, msg);
                    }
                } catch (InterruptedException _) {
                    Thread.currentThread().interrupt();
                    return;
                }
            }
        });
    }

    public void stop() {
        stopped = true;
        if (readerThread != null) {
            readerThread.interrupt();
        }
    }

    public List<QueueMessage> flush() {
        synchronized (mutex) {
            if (messages.isEmpty()) return null;

            List<QueueMessage> flushed = new ArrayList<>(messages);
            messages.clear();
            return flushed;
        }
    }

    private void append(int from, Message message) {
        synchronized (mutex) {
            messages.add(new QueueMessage(from, message));
        }
    }
}
