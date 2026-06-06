package com.hornflakes.gorunsrestwalk.gameloop;

import static com.hornflakes.gorunsrestwalk.gameloop.Spec.*;

public final class Bullet {

    public final Rect rect;
    public final double vx;
    public final double vy;

    private Bullet(double x, double y, double vx, double vy) {
        this.rect = new Rect(x, y, BULLET_WIDTH, BULLET_HEIGHT);
        this.vx = vx;
        this.vy = vy;
    }

    public static Bullet fromPlayer(Player player) {
        double x;
        if (player.dir[0] == 1) {
            x = player.rect.x + player.rect.width + 1;
        } else {
            x = player.rect.x - BULLET_WIDTH - 1;
        }

        return new Bullet(
                x, 0,
                player.dir[0] * BULLET_SPEED_MS,
                player.dir[1] * BULLET_SPEED_MS
        );
    }

    public void updatePosition(double deltaMs) {
        rect.x += vx * deltaMs;
        rect.y += vy * deltaMs;
    }
}
