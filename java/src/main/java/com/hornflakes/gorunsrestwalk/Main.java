package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.PairingServer;
import com.hornflakes.gorunsrestwalk.server.Socket;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.ServerConnector;
import org.eclipse.jetty.websocket.server.WebSocketUpgradeHandler;

public class Main {

    public static void main(String[] args) throws Exception {
        PairingServer pairing = new PairingServer();

        pairing.setOnPair((s0, s1) -> {
            Logger log = Logger.forPair(s0.playerId(), s1.playerId());
            log.logMilestone("websockets paired", "");
        });

        Server server = new Server();

        ServerConnector connector = new ServerConnector(server);
        connector.setPort(37373);
        server.addConnector(connector);

        WebSocketUpgradeHandler wsHandler = WebSocketUpgradeHandler.from(server, container -> {
            container.addMapping("/", (req, resp, cb) -> {
                Socket socket = pairing.createSocket();
                socket.setOnReady(() -> pairing.register(socket));
                return socket;
            });
        });

        server.setHandler(wsHandler);
        server.start();

        Log.info("server listening", "addr=:37373");

        server.join();
    }
}
