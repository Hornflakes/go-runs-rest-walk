package com.hornflakes.gorunsrestwalk;

import com.hornflakes.gorunsrestwalk.logger.Logger;
import io.netty.bootstrap.ServerBootstrap;
import io.netty.channel.ChannelInitializer;
import io.netty.channel.nio.NioEventLoopGroup;
import io.netty.channel.socket.SocketChannel;
import io.netty.channel.socket.nio.NioServerSocketChannel;
import io.netty.handler.codec.http.HttpObjectAggregator;
import io.netty.handler.codec.http.HttpServerCodec;

public class Main {

    private static final int PORT = 37373;

    public static void main(String[] args) {
        boolean verbose = false;
        for (String arg : args) {
            if ("-verbose".equals(arg)) {
                verbose = true;
            }
        }

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
