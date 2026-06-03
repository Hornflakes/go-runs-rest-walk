import { Socket } from './socket.js';
import { MessageType, createMessage, createWinnerMessage, createLoserMessage } from './message.js';
import { Logger } from './logger.js';
import { Queue } from './queue.js';
import { Player, Bullet, createBulletFromPlayer } from './objects.js';
import { GameFrameStats, activeGames, addActiveGame, removeActiveGame } from './stats.js';
import {
    Player0Spawn,
    Player1Spawn,
    Player0Dir,
    Player1Dir,
    Player0FireRateMs,
    Player1FireRateMs,
    BulletSpeedMs,
    TickTargetMicros,
    MicrosPerMs,
} from './spec.js';
import * as logger from './logger.js';

function nowMicros(): number {
    return Math.floor(performance.now() * 1000);
}

function nowMillis(): number {
    return performance.now();
}

export function runGame(s0: Socket, s1: Socket, verbose: boolean): void {
    const sockets: [Socket, Socket] = [s0, s1];
    const players: [Player, Player] = [
        new Player(Player0Spawn, Player0Dir, Player0FireRateMs),
        new Player(Player1Spawn, Player1Dir, Player1FireRateMs),
    ];
    const bullets: Bullet[] = [];
    const stats = new GameFrameStats();
    const queue = new Queue();
    const log = new Logger(s0.playerId, s1.playerId);

    queue.start(s0, s1);
    addActiveGame();

    s0.send(createMessage(MessageType.GameOn));
    s1.send(createMessage(MessageType.GameOn));
    log.info('game on', '');

    let ticks = 0;
    const tickStartTime = nowMicros();
    let lastLoopTime = nowMicros();

    function tick(): void {
        ticks++;

        const startTime = nowMicros();
        const deltaTime = startTime - lastLoopTime;

        if (ticks > 1) {
            stats.addDeltaTime(deltaTime);
        }

        updateStateFromMessageQueue();
        updateBulletsPositions(deltaTime);
        checkBulletBulletCollisions();

        const loser = checkBulletPlayerCollisions();
        if (loser !== null) {
            const winner = getOtherPlayer(loser);
            const winnerSocket = getPlayerSocket(winner);
            const loserSocket = getPlayerSocket(loser);
            const winnerId = winnerSocket.playerId;

            winnerSocket.send(createWinnerMessage(winnerId, stats.toString(), activeGames));
            loserSocket.send(createLoserMessage());

            winnerSocket.close();
            loserSocket.close();

            queue.stop();
            removeActiveGame();

            const elapsedUs = nowMicros() - tickStartTime;
            log.milestone(
                'game over',
                `winner=${winnerId} histogram=${stats} active_games=${activeGames} ticks=${ticks} elapsed=${formatElapsed(elapsedUs)} bullets=${bullets.length}`,
            );
            return;
        }

        const afterTime = nowMicros();
        const sleepUs = TickTargetMicros - (afterTime - startTime);
        const sleepMs = Math.max(0, sleepUs / 1000);

        setTimeout(tick, sleepMs);
        lastLoopTime = startTime;
    }

    function updateStateFromMessageQueue(): void {
        const messages = queue.flush();
        if (messages === null) return;

        for (const qm of messages) {
            if (qm.message.type === MessageType.Shoot) {
                const player = players[qm.from - 1];
                const fired = player.fire(nowMillis());

                if (fired) {
                    const bullet = createBulletFromPlayer(player, BulletSpeedMs);
                    bullets.push(bullet);

                    if (verbose) {
                        log.info(
                            'player shot',
                            `${logger.player(sockets[qm.from - 1].playerId)} bullet=${bullets.length}`,
                        );
                    }
                }
            }
        }
    }

    function updateBulletsPositions(deltaTime: number): void {
        const deltaMs = deltaTime / MicrosPerMs;
        for (const bullet of bullets) {
            bullet.rect.x += deltaMs * bullet.velocity[0];
            bullet.rect.y += deltaMs * bullet.velocity[1];
        }
    }

    function checkBulletBulletCollisions(): void {
        for (let i = 0; i < bullets.length; i++) {
            for (let j = i + 1; j < bullets.length; j++) {
                if (bullets[i].rect.collides(bullets[j].rect)) {
                    bullets.splice(j, 1);
                    bullets.splice(i, 1);
                    return;
                }
            }
        }
    }

    function checkBulletPlayerCollisions(): Player | null {
        for (const player of players) {
            for (const bullet of bullets) {
                if (bullet.rect.collides(player.rect)) {
                    return player;
                }
            }
        }
        return null;
    }

    function getPlayerSocket(player: Player): Socket {
        if (player === players[0]) return sockets[0];
        return sockets[1];
    }

    function getOtherPlayer(player: Player): Player {
        if (player === players[0]) return players[1];
        return players[0];
    }

    function formatElapsed(micros: number): string {
        const ms = micros / 1000;
        if (ms < 1000) return `${ms.toFixed(3)}ms`;
        const s = ms / 1000;
        return `${s.toFixed(3)}s`;
    }

    tick();
}
