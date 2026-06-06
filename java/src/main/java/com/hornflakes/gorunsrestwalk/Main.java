package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.gameloop.WaitForReady;
import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.Server;
import com.hornflakes.gorunsrestwalk.server.WebSocketHandler;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.MultiThreadIoEventLoopGroup;
import io.netty.channel.nio.NioIoHandler;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;
import io.netty.handler.codec.http.websocketx.WebSocketServerProtocolHandler;

public class Main {

    private static final int PORT = 37373;

    public static void main(String[] args) {
        boolean verbose = false;
        for (String arg : args) {
            if ("-verbose".equals(arg)) {
                verbose = true;
            }
        }

        var srv = new Server(pair -> {
            var log = Logger.forPair(pair[0].playerId(), pair[1].playerId());
            log.milestone("websockets paired", "");

            Thread.startVirtualThread(() -> {
                boolean ok = WaitForReady.execute(pair[0], pair[1]);

                if (!ok) {
                    log.warn("websockets ready handshake failed", "");
                    pair[0].close();
                    pair[1].close();
                    return;
                }

                log.milestone("websockets ready handshake ok", "");
                // Phase 5: game loop
            });
        });

        var bossGroup = new MultiThreadIoEventLoopGroup(1, NioIoHandler.newFactory());
        var workerGroup = new MultiThreadIoEventLoopGroup(NioIoHandler.newFactory());

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
                            pipeline.addLast(new WebSocketHandler(srv::registerSocket));
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
