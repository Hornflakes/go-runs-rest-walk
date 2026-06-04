package com.hornflakes.gorunsrestwalk;

import org.eclipse.jetty.websocket.api.Session;
import org.eclipse.jetty.websocket.api.annotations.OnWebSocketOpen;
import org.eclipse.jetty.websocket.api.annotations.WebSocket;

@WebSocket
public class PlaceholderEndpoint {

    @OnWebSocketOpen
    public void onOpen(Session session) {
    }
}
