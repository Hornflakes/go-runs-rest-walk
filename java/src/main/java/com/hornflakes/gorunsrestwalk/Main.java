package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.gameloop.Game;
import com.hornflakes.gorunsrestwalk.gameloop.WaitForReady;
import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.PairingServer;
import com.hornflakes.gorunsrestwalk.server.Socket;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.ServerConnector;
import org.eclipse.jetty.websocket.server.WebSocketUpgradeHandler;

import java.time.Duration;

public class Main {
    
    private static final Duration WEBSOCKET_IDLE_TIMEOUT = Duration.ZERO;

    public static void main(String[] args) throws Exception {
        boolean verbose = java.util.Arrays.asList(args).contains("--verbose");

        PairingServer pairing = new PairingServer();

        pairing.setOnPair((s0, s1) -> {
            Logger log = Logger.forPair(s0.playerId(), s1.playerId());
            log.logMilestone("websockets paired", "");

            Thread.ofVirtual().start(() -> {
                boolean ok = WaitForReady.await(s0, s1);

                if (!ok) {
                    log.logWarn("websockets ready handshake failed", "");
                    s0.close();
                    s1.close();
                    return;
                }

                log.logMilestone("websockets ready handshake ok", "");

                new Game(s0, s1, verbose).run();
            });
        });

        Server server = new Server();

        ServerConnector connector = new ServerConnector(server);
        connector.setPort(37373);
        connector.setIdleTimeout(0);
        server.addConnector(connector);

        WebSocketUpgradeHandler wsHandler = WebSocketUpgradeHandler.from(server, container -> {
            container.setIdleTimeout(WEBSOCKET_IDLE_TIMEOUT);
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
