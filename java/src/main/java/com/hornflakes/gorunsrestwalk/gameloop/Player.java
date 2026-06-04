package com.hornflakes.gorunsrestwalk.gameloop;

public class Player {

    public final Rect rect;
    public final double[] dir;
    public final long fireRate;
    private long lastFireTime;

    public Player(double[] pos, double[] dir, long fireRate) {
        this.rect = new Rect(pos[0], pos[1], Spec.PLAYER_WIDTH, Spec.PLAYER_HEIGHT);
        this.dir = dir;
        this.fireRate = fireRate;
        this.lastFireTime = 0;
    }

    public boolean fire() {
        long now = System.currentTimeMillis();

        if (fireRate > now - lastFireTime) {
            return false;
        }

        lastFireTime = now;
        return true;
    }
}
