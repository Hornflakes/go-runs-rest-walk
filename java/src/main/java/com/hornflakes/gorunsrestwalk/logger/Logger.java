package com.hornflakes.gorunsrestwalk.logger;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

public final class Logger {

    public enum Level {
        MILESTONE("\u001b[32m"),
        INFO("\u001b[36m"),
        WARN("\u001b[33m"),
        SOFT_ERROR("\u001b[35m"),
        HARD_ERROR("\u001b[31m");

        private final String ansi;

        Level(String ansi) {
            this.ansi = ansi;
        }

        String color(String event) {
            return ansi + event + "\u001b[0m";
        }
    }

    private static final DateTimeFormatter FMT = DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss");

    private final String prefix;

    private Logger(String prefix) {
        this.prefix = prefix;
    }

    public static Logger forPair(long playerId0, long playerId1) {
        return new Logger(pairPrefix(playerId0, playerId1));
    }

    private static String formatPrefix(String prefix) {
        if (prefix.isEmpty()) return "";
        return "[" + prefix + "] ";
    }

    static void write(String prefix, Level level, String event, String detail) {
        String ts = FMT.format(LocalDateTime.now());
        String p = formatPrefix(prefix);
        String colored = level.color(event);

        if (detail.isEmpty()) {
            System.out.println(ts + " " + p + colored);
        } else {
            System.out.println(ts + " " + p + colored + " | " + detail);
        }
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

    public void logMilestone(String event, String detail) { write(prefix, Level.MILESTONE, event, detail); }
    public void logInfo(String event, String detail)      { write(prefix, Level.INFO, event, detail); }
    public void logWarn(String event, String detail)      { write(prefix, Level.WARN, event, detail); }
    public void logSoftError(String event, String detail) { write(prefix, Level.SOFT_ERROR, event, detail); }
    public void logHardError(String event, String detail) { write(prefix, Level.HARD_ERROR, event, detail); }
}
