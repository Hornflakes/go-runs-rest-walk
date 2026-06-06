package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;

import java.util.concurrent.TimeUnit;

public final class WaitForReady {

    private static final long INBOUND_POLL_MS = 50;

    private WaitForReady() {}

    public static boolean await(Socket s0, Socket s1) {
        s0.send(Message.createReady(s1.playerId()));
        s1.send(Message.createReady(s0.playerId()));

        long deadlineNanos = System.nanoTime() + TimeUnit.MILLISECONDS.toNanos(Spec.READY_TIMEOUT_MS);

        int count = 0;
        boolean in0Open = true;
        boolean in1Open = true;

        try {
            while (count < 2) {
                if (s0.disconnected() || s0.closed() || s1.disconnected() || s1.closed()) {
                    return false;
                }
                if (System.nanoTime() >= deadlineNanos) {
                    return false;
                }

                if (in0Open) {
                    Message msg = s0.pollInbound(INBOUND_POLL_MS, TimeUnit.MILLISECONDS);
                    if (msg != null) {
                        if (Socket.isInboundClosed(msg)) {
                            in0Open = false;
                        } else if (msg.type() == Message.READY) {
                            count++;
                            in0Open = false;
                        }
                    }
                }

                if (in1Open) {
                    Message msg = s1.pollInbound(INBOUND_POLL_MS, TimeUnit.MILLISECONDS);
                    if (msg != null) {
                        if (Socket.isInboundClosed(msg)) {
                            in1Open = false;
                        } else if (msg.type() == Message.READY) {
                            count++;
                            in1Open = false;
                        }
                    }
                }

                if (!in0Open && !in1Open && count < 2) {
                    return false;
                }
            }
        } catch (InterruptedException _) {
            Thread.currentThread().interrupt();
            return false;
        }

        return count == 2;
    }
}
