package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.ServerConnector;
import org.eclipse.jetty.websocket.server.WebSocketUpgradeHandler;

import java.util.concurrent.atomic.AtomicLong;

public class Main {

    private static final AtomicLong nextPlayerId = new AtomicLong(0);

    public static void main(String[] args) throws Exception {
        Server server = new Server();

        ServerConnector connector = new ServerConnector(server);
        connector.setPort(37373);
        server.addConnector(connector);

        WebSocketUpgradeHandler wsHandler = WebSocketUpgradeHandler.from(server, container -> {
            container.addMapping("/", (req, resp, cb) -> {
                Socket socket = new Socket();
                socket.setPlayerId(nextPlayerId.incrementAndGet());
                return socket;
            });
        });

        server.setHandler(wsHandler);
        server.start();

        Log.info("server listening", "addr=:37373");

        server.join();
    }
}
