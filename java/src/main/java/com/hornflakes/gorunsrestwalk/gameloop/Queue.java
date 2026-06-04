package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;

import java.util.ArrayList;
import java.util.List;

public class Queue {

    private final List<QueueMessage> messages = new ArrayList<>();
    private final Object mutex = new Object();
    private volatile boolean stopped;

    public record QueueMessage(int from, Message message) {}

    public void start(Socket s0, Socket s1) {
        Thread.ofVirtual().start(() -> readLoop(s0, 1));
        Thread.ofVirtual().start(() -> readLoop(s1, 2));
    }

    private void readLoop(Socket socket, int from) {
        while (!stopped) {
            try {
                Message msg = socket.in().poll(1, java.util.concurrent.TimeUnit.SECONDS);
                if (msg == null) continue;

                synchronized (mutex) {
                    messages.add(new QueueMessage(from, msg));
                }
            } catch (InterruptedException _) {
                return;
            }
        }
    }

    public void stop() {
        stopped = true;
    }

    public List<QueueMessage> flush() {
        synchronized (mutex) {
            if (messages.isEmpty()) {
                return null;
            }

            List<QueueMessage> flushed = new ArrayList<>(messages);
            messages.clear();
            return flushed;
        }
    }
}
