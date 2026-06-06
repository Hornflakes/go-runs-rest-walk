package com.hornflakes.gorunsrestwalk.server;

import com.hornflakes.gorunsrestwalk.logger.Logger;

import java.util.concurrent.atomic.AtomicLong;
import java.util.function.Consumer;

public class PairingServer {

    private final Consumer<Socket[]> onPaired;
    private final AtomicLong nextPlayerId = new AtomicLong(0);
    private final Object lock = new Object();
    private Socket waitingSocket;

    public PairingServer(Consumer<Socket[]> onPaired) {
        this.onPaired = onPaired;
    }

    public void registerSocket(Socket socket) {
        Socket[] pair = null;

        synchronized (lock) {
            socket.setPlayerId(nextPlayerId.incrementAndGet());

            if (socketAlive(waitingSocket)) {
                pair = new Socket[]{waitingSocket, socket};
                waitingSocket = null;
            } else {
                if (waitingSocket != null) {
                    waitingSocket.close();
                    waitingSocket = null;
                }
                waitingSocket = socket;
            }
        }

        socket.send(Message.createHello(socket.playerId()));
        Logger.GLOBAL.info("websocket connected",
                Logger.playerWithAddr(socket.playerId(), socket.remoteAddr()));

        if (pair != null) {
            onPaired.accept(pair);
        }
    }

    private static boolean socketAlive(Socket s) {
        return s != null && !s.closed() && !s.disconnected();
    }

    public static void watchPairDisconnect(Socket s0, Socket s1, Runnable onDisconnect) {
        Thread.startVirtualThread(() -> {
            while (true) {
                if (s0.disconnected() || s0.closed() || s1.disconnected() || s1.closed()) {
                    onDisconnect.run();
                    return;
                }
                try {
                    Thread.sleep(50);
                } catch (InterruptedException e) {
                    return;
                }
            }
        });
    }
}
