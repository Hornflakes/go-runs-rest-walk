package com.hornflakes.gorunsrestwalk.logger;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

public final class Logger {

    private static final DateTimeFormatter FMT = DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss");

    private static final String RESET = "\033[0m";
    private static final String GREEN = "\033[32m";
    private static final String CYAN = "\033[36m";
    private static final String YELLOW = "\033[33m";
    private static final String MAGENTA = "\033[35m";
    private static final String RED = "\033[31m";

    public static final Logger GLOBAL = new Logger("");

    private final String prefix;

    private Logger(String prefix) {
        this.prefix = prefix;
    }

    public static Logger forPair(long playerId0, long playerId1) {
        return new Logger(pairPrefix(playerId0, playerId1));
    }

    public static String player(long id) {
        return "player=" + id;
    }

    public static String playerWithAddr(long id, String addr) {
        return "player=" + id + " addr=" + addr;
    }

    public static String playerPair(long playerId0, long playerId1) {
        return player(playerId0) + " vs " + player(playerId1);
    }

    public static String pairPrefix(long playerId0, long playerId1) {
        return playerId0 + " vs " + playerId1;
    }


    public void milestone(String event, String detail) {
        write(prefix, GREEN, event, detail);
    }

    public void info(String event, String detail) {
        write(prefix, CYAN, event, detail);
    }

    public void warn(String event, String detail) {
        write(prefix, YELLOW, event, detail);
    }

    public void softError(String event, String detail) {
        write(prefix, MAGENTA, event, detail);
    }

    public void hardError(String event, String detail) {
        write(prefix, RED, event, detail);
    }

    private static void write(String prefix, String color, String event, String detail) {
        var sb = new StringBuilder();
        sb.append(LocalDateTime.now().format(FMT)).append(' ');

        if (!prefix.isEmpty()) {
            sb.append('[').append(prefix).append("] ");
        }

        sb.append(color).append(event).append(RESET);

        if (detail != null && !detail.isEmpty()) {
            sb.append(" | ").append(detail);
        }

        System.out.println(sb);
    }
}
