package com.hornflakes.gorunsrestwalk.gameloop;

import static com.hornflakes.gorunsrestwalk.gameloop.Spec.*;

public final class Player {

    public final Rect rect;
    public final double[] dir;
    public final long fireRateMs;
    private long lastFireTime;

    public Player(double[] spawn, double[] dir, long fireRateMs) {
        this.rect = new Rect(spawn[0], spawn[1], PLAYER_WIDTH, PLAYER_HEIGHT);
        this.dir = dir;
        this.fireRateMs = fireRateMs;
    }

    public boolean fire() {
        long now = System.currentTimeMillis();
        if (fireRateMs > now - lastFireTime) {
            return false;
        }
        lastFireTime = now;
        return true;
    }
}
