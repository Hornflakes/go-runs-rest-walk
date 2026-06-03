import { Rect, Vector2D } from './geometry.js';
import { PlayerWidth, PlayerHeight, BulletWidth, BulletHeight } from './spec.js';

export class Player {
    rect: Rect;
    dir: Vector2D;
    fireRate: number;
    private lastFireTime: number = 0;

    constructor(pos: Vector2D, dir: Vector2D, fireRate: number) {
        this.rect = new Rect(pos[0], pos[1], PlayerWidth, PlayerHeight);
        this.dir = dir;
        this.fireRate = fireRate;
    }

    fire(nowMs: number): boolean {
        if (this.fireRate > nowMs - this.lastFireTime) {
            return false;
        }
        this.lastFireTime = nowMs;
        return true;
    }
}

export class Bullet {
    rect: Rect;
    velocity: Vector2D;

    constructor() {
        this.rect = new Rect(0, 0, BulletWidth, BulletHeight);
        this.velocity = [0, 0];
    }
}

export function createBulletFromPlayer(player: Player, speedMs: number): Bullet {
    const bullet = new Bullet();

    if (player.dir[0] === 1) {
        bullet.rect.setPosition(player.rect.x + player.rect.width + 1, 0);
    } else {
        bullet.rect.setPosition(player.rect.x - BulletWidth - 1, 0);
    }

    bullet.velocity[0] = player.dir[0] * speedMs;
    bullet.velocity[1] = player.dir[1] * speedMs;

    return bullet;
}
