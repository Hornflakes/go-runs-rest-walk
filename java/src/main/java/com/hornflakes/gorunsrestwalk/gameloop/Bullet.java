package com.hornflakes.gorunsrestwalk.gameloop;

public class Bullet {

    public final Rect rect;
    public final double[] velocity;

    private Bullet() {
        this.rect = new Rect(0, 0, Spec.BULLET_WIDTH, Spec.BULLET_HEIGHT);
        this.velocity = new double[]{0, 0};
    }

    public static Bullet fromPlayer(Player player, double speedMs) {
        Bullet bullet = new Bullet();

        if (player.dir[0] == 1) {
            bullet.rect.setPosition(player.rect.x + player.rect.width + 1, 0);
        } else {
            bullet.rect.setPosition(player.rect.x - Spec.BULLET_WIDTH - 1, 0);
        }

        bullet.velocity[0] = player.dir[0] * speedMs;
        bullet.velocity[1] = player.dir[1] * speedMs;

        return bullet;
    }
}
