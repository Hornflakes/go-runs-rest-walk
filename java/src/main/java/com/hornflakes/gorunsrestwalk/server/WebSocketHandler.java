package com.hornflakes.gorunsrestwalk.server;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.websocketx.TextWebSocketFrame;
import io.netty.handler.codec.http.websocketx.WebSocketServerProtocolHandler;

import java.util.function.Consumer;

public class WebSocketHandler extends SimpleChannelInboundHandler<TextWebSocketFrame> {

    private final Consumer<Socket> onNewSocket;
    private Socket socket;

    public WebSocketHandler(Consumer<Socket> onNewSocket) {
        this.onNewSocket = onNewSocket;
    }

    @Override
    public void userEventTriggered(ChannelHandlerContext ctx, Object evt) throws Exception {
        if (evt instanceof WebSocketServerProtocolHandler.HandshakeComplete) {
            socket = new Socket(ctx.channel());
            onNewSocket.accept(socket);
        } else {
            super.userEventTriggered(ctx, evt);
        }
    }

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, TextWebSocketFrame frame) {
        if (socket == null) return;

        Message msg;
        try {
            msg = Message.unmarshal(frame.text());
        } catch (JsonProcessingException e) {
            Logger.GLOBAL.softError("websocket message unmarshal failed", socket.logDetail(e));
            return;
        }

        socket.pushMessage(msg);
    }

    @Override
    public void channelInactive(ChannelHandlerContext ctx) throws Exception {
        if (socket != null) {
            socket.markDisconnected();
        }
        super.channelInactive(ctx);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        if (socket != null && !socket.disconnected()) {
            Logger.GLOBAL.warn("websocket read ended", socket.logDetail(cause));
        }
        ctx.close();
    }
}
