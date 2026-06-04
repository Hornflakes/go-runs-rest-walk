package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;

public final class WaitForReady {

    private static final long READY_TIMEOUT_MS = 30_000;
    private static final long DISCONNECT_POLL_MS = 50;

    private WaitForReady() {}

    public static boolean await(Socket s0, Socket s1) {
        s0.send(Message.createReady(s1.playerId()));
        s1.send(Message.createReady(s0.playerId()));

        var latch = new CountDownLatch(2);
        var failed = new AtomicBoolean(false);

        Thread t0 = Thread.ofVirtual().start(() -> {
            if (waitForReadyMessage(s0, failed)) {
                latch.countDown();
            } else {
                failed.set(true);
            }
        });

        Thread t1 = Thread.ofVirtual().start(() -> {
            if (waitForReadyMessage(s1, failed)) {
                latch.countDown();
            } else {
                failed.set(true);
            }
        });

        Thread watchdog = Thread.ofVirtual().start(() -> {
            while (!failed.get()) {
                if (s0.disconnected() || s0.closed() || s1.disconnected() || s1.closed()) {
                    failed.set(true);
                    t0.interrupt();
                    t1.interrupt();
                    return;
                }
                try {
                    Thread.sleep(DISCONNECT_POLL_MS);
                } catch (InterruptedException _) {
                    return;
                }
            }
        });

        try {
            boolean ok = latch.await(READY_TIMEOUT_MS, TimeUnit.MILLISECONDS);
            if (!ok || failed.get()) {
                failed.set(true);
                t0.interrupt();
                t1.interrupt();
                watchdog.interrupt();
                return false;
            }
            watchdog.interrupt();
            return true;
        } catch (InterruptedException _) {
            failed.set(true);
            t0.interrupt();
            t1.interrupt();
            watchdog.interrupt();
            return false;
        }
    }

    private static boolean waitForReadyMessage(Socket socket, AtomicBoolean failed) {
        while (!failed.get()) {
            if (socket.disconnected() || socket.closed()) return false;

            try {
                Message msg = socket.in().poll(1, TimeUnit.SECONDS);
                if (msg == null) continue;
                if (msg.type() == Message.READY) return true;
            } catch (InterruptedException _) {
                return false;
            }
        }
        return false;
    }
}
