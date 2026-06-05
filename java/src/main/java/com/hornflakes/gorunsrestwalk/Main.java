package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;
import com.hornflakes.gorunsrestwalk.server.WebSocketHandler;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.codec.http.websocketx.WebSocketServerProtocolHandler;

import java.util.concurrent.atomic.AtomicLong;

public class Main {

    private static final int PORT = 37373;

    public static void main(String[] args) {
        boolean verbose = false;
        for (String arg : args) {
            if ("-verbose".equals(arg)) {
                verbose = true;
            }
        }

        var nextPlayerId = new AtomicLong(0);

        var bossGroup = new NioEventLoopGroup(1);
        var workerGroup = new NioEventLoopGroup();

        try {
            var bootstrap = new ServerBootstrap();
            bootstrap.group(bossGroup, workerGroup)
                    .channel(NioServerSocketChannel.class)
                    .childHandler(new ChannelInitializer<SocketChannel>() {
                        @Override
                        protected void initChannel(SocketChannel ch) {
                            var pipeline = ch.pipeline();
                            pipeline.addLast(new HttpServerCodec());
                            pipeline.addLast(new HttpObjectAggregator(65536));
                            pipeline.addLast(new WebSocketServerProtocolHandler("/"));
                            pipeline.addLast(new WebSocketHandler(socket -> {
                                socket.setPlayerId(nextPlayerId.incrementAndGet());
                                socket.send(Message.createHello(socket.playerId()));
                                Logger.GLOBAL.info("websocket connected",
                                        Logger.playerWithAddr(socket.playerId(), socket.remoteAddr()));
                            }));
                        }
                    });

            var future = bootstrap.bind(PORT).sync();

            Logger.GLOBAL.info("server listening", "addr=:" + PORT);

            future.channel().closeFuture().sync();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        } finally {
            bossGroup.shutdownGracefully();
            workerGroup.shutdownGracefully();
        }
    }
}
