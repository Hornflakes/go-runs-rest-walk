package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.logger.Log;
import com.hornflakes.gorunsrestwalk.server.Message;
import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.server.ServerConnector;
import org.eclipse.jetty.websocket.server.WebSocketUpgradeHandler;

public class Main {

    public static void main(String[] args) throws Exception {
        Server server = new Server();

        ServerConnector connector = new ServerConnector(server);
        connector.setPort(37373);
        server.addConnector(connector);

        WebSocketUpgradeHandler wsHandler = WebSocketUpgradeHandler.from(server, container -> {
            container.addMapping("/", (req, resp, cb) -> new PlaceholderEndpoint());
        });

        server.setHandler(wsHandler);
        server.start();

        Log.info("server listening", "addr=:37373");

        server.join();
    }
}
