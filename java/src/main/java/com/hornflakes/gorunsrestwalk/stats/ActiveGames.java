package com.hornflakes.gorunsrestwalk.stats;

public final class ActiveGames {

    private ActiveGames() {}

    private static final Object mutex = new Object();
    private static long count;

    public static void add() {
        synchronized (mutex) {
            count++;
        }
    }

    public static void remove() {
        synchronized (mutex) {
            count--;
        }
    }

    public static long get() {
        synchronized (mutex) {
            return count;
        }
    }
}
