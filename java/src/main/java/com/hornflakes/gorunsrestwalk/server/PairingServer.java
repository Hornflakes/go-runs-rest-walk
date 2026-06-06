package com.hornflakes.gorunsrestwalk.server;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicLong;
import java.util.function.BiConsumer;

public class PairingServer {

    private final AtomicLong nextPlayerId = new AtomicLong(0);
    private final Object mutex = new Object();
    private final BlockingQueue<Socket[]> pairOut = new LinkedBlockingQueue<>(4);
    private Socket waitingSocket;
    private BiConsumer<Socket, Socket> onPair;

    public PairingServer() {
        Thread.ofVirtual().start(this::dispatchPairs);
    }

    public void setOnPair(BiConsumer<Socket, Socket> onPair) {
        this.onPair = onPair;
    }

    public Socket createSocket() {
        Socket socket = new Socket();
        socket.setPlayerId(nextPlayerId.incrementAndGet());
        return socket;
    }

    public void register(Socket socket) {
        Socket[] pair = null;

        synchronized (mutex) {
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

        if (pair != null) {
            try {
                pairOut.put(pair);
            } catch (InterruptedException _) {
                Thread.currentThread().interrupt();
                pair[0].close();
                pair[1].close();
                return;
            }
        }

        socket.send(Message.createHello(socket.playerId()));
        Log.info("websocket connected", Logger.playerWithAddr(socket.playerId(), socket.remoteAddr()));
    }

    private void dispatchPairs() {
        while (true) {
            try {
                Socket[] pair = pairOut.take();
                if (onPair != null) {
                    onPair.accept(pair[0], pair[1]);
                }
            } catch (InterruptedException _) {
                Thread.currentThread().interrupt();
                return;
            }
        }
    }

    private static boolean socketAlive(Socket s) {
        return s != null && !s.closed() && !s.disconnected();
    }
}
