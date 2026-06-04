package com.hornflakes.gorunsrestwalk.server;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;

import java.util.concurrent.atomic.AtomicLong;
import java.util.function.BiConsumer;

public class PairingServer {

    private final AtomicLong nextPlayerId = new AtomicLong(0);
    private final Object mutex = new Object();
    private Socket waitingSocket;
    private BiConsumer<Socket, Socket> onPair;

    public void setOnPair(BiConsumer<Socket, Socket> onPair) {
        this.onPair = onPair;
    }

    public Socket createSocket() {
        Socket socket = new Socket();
        socket.setPlayerId(nextPlayerId.incrementAndGet());
        return socket;
    }

    public void register(Socket socket) {
        synchronized (mutex) {
            if (socketAlive(waitingSocket)) {
                Socket paired = waitingSocket;
                waitingSocket = null;

                if (onPair != null) {
                    onPair.accept(paired, socket);
                }
                return;
            }

            if (waitingSocket != null) {
                waitingSocket.close();
                waitingSocket = null;
            }

            waitingSocket = socket;
        }
    }

    private static boolean socketAlive(Socket s) {
        return s != null && !s.closed() && !s.disconnected();
    }
}
