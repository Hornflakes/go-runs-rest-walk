package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

public final class WaitForReady {

    private static final long TIMEOUT_MS = 30_000;

    public static boolean execute(Socket s0, Socket s1) {
        s0.send(Message.createReady(s1.playerId()));
        s1.send(Message.createReady(s0.playerId()));

        var latch = new CountDownLatch(2);
        var failed = new AtomicBoolean(false);

        var t0 = Thread.startVirtualThread(() -> awaitReady(s0, latch, failed));
        var t1 = Thread.startVirtualThread(() -> awaitReady(s1, latch, failed));

        var watcher = Thread.startVirtualThread(() -> {
            try {
                while (!Thread.currentThread().isInterrupted()) {
                    if (s0.disconnected() || s0.closed() || s1.disconnected() || s1.closed()) {
                        failed.set(true);
                        t0.interrupt();
                        t1.interrupt();
                        return;
                    }
                    Thread.sleep(50);
                }
            } catch (InterruptedException e) {
                // stopped
            }
        });

        try {
            if (!latch.await(TIMEOUT_MS, TimeUnit.MILLISECONDS)) {
                failed.set(true);
                t0.interrupt();
                t1.interrupt();
            }
        } catch (InterruptedException e) {
            failed.set(true);
            t0.interrupt();
            t1.interrupt();
        }

        watcher.interrupt();
        return !failed.get();
    }

    private static void awaitReady(Socket s, CountDownLatch latch, AtomicBoolean failed) {
        try {
            while (!failed.get()) {
                var msg = s.in().poll(1, TimeUnit.SECONDS);
                if (msg == null) continue;
                if (msg.getType() == Message.READY) {
                    latch.countDown();
                    return;
                }
            }
        } catch (InterruptedException e) {
            // cancelled by timeout or disconnect
        }
    }
}
