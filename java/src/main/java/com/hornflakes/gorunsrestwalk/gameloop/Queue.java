package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentLinkedQueue;

public final class Queue {

    public record QueueMessage(int from, Message message) {}

    private final ConcurrentLinkedQueue<QueueMessage> messages = new ConcurrentLinkedQueue<>();
    private Thread reader0;
    private Thread reader1;

    public void start(Socket s0, Socket s1) {
        reader0 = Thread.startVirtualThread(() -> readLoop(s0, 1));
        reader1 = Thread.startVirtualThread(() -> readLoop(s1, 2));
    }

    public void stop() {
        if (reader0 != null) reader0.interrupt();
        if (reader1 != null) reader1.interrupt();
    }

    public List<QueueMessage> flush() {
        var batch = new ArrayList<QueueMessage>();
        QueueMessage msg;
        while ((msg = messages.poll()) != null) {
            batch.add(msg);
        }
        return batch.isEmpty() ? null : batch;
    }

    private void readLoop(Socket socket, int from) {
        try {
            while (!Thread.currentThread().isInterrupted()) {
                var msg = socket.in().take();
                messages.add(new QueueMessage(from, msg));
            }
        } catch (InterruptedException e) {
            // stopped
        }
    }
}
