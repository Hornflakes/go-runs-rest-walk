package com.hornflakes.gorunsrestwalk.logger;

public final class Log {

    private Log() {}

    public static void milestone(String event, String detail) {
        Logger.write("", Logger.Level.MILESTONE, event, detail);
    }

    public static void info(String event, String detail) {
        Logger.write("", Logger.Level.INFO, event, detail);
    }

    public static void warn(String event, String detail) {
        Logger.write("", Logger.Level.WARN, event, detail);
    }

    public static void softError(String event, String detail) {
        Logger.write("", Logger.Level.SOFT_ERROR, event, detail);
    }

    public static void hardError(String event, String detail) {
        Logger.write("", Logger.Level.HARD_ERROR, event, detail);
    }
}
