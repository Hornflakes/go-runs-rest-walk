package com.hornflakes.gorunsrestwalk.stats;

import java.util.concurrent.atomic.AtomicInteger;

public final class ActiveGames {

    private static final AtomicInteger count = new AtomicInteger(0);

    public static int increment() { return count.incrementAndGet(); }
    public static int decrement() { return count.decrementAndGet(); }
    public static int get() { return count.get(); }
}
