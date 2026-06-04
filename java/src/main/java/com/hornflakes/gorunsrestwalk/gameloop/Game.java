package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;
import com.hornflakes.gorunsrestwalk.stats.ActiveGames;
import com.hornflakes.gorunsrestwalk.stats.GameFrameStats;

import java.time.Duration;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

public class Game {

    private final Socket[] sockets;
    private final Player[] players;
    private final List<Bullet> bullets = new ArrayList<>();
    private final GameFrameStats stats = new GameFrameStats();
    private final boolean verbose;
    private Queue queue;

    public Game(Socket s0, Socket s1, boolean verbose) {
        this.sockets = new Socket[]{s0, s1};
        this.players = new Player[]{
            new Player(Spec.PLAYER0_SPAWN, Spec.PLAYER0_DIR, Spec.PLAYER0_FIRE_RATE_MS),
            new Player(Spec.PLAYER1_SPAWN, Spec.PLAYER1_DIR, Spec.PLAYER1_FIRE_RATE_MS),
        };
        this.verbose = verbose;
    }

    private void start() {
        queue = new Queue();
        queue.start(sockets[0], sockets[1]);
    }

    private void stop() {
        if (queue != null) {
            queue.stop();
        }
    }

    private void updateStateFromMessageQueue(Logger log) {
        List<Queue.QueueMessage> messages = queue.flush();
        if (messages == null) return;

        for (Queue.QueueMessage qm : messages) {
            if (qm.message().type() == Message.SHOOT) {
                Player player = players[qm.from() - 1];
                boolean fired = player.fire();

                if (fired) {
                    Bullet bullet = Bullet.fromPlayer(player, Spec.BULLET_SPEED_MS);
                    bullets.add(bullet);

                    if (verbose) {
                        log.logInfo("player shot",
                            Logger.player(sockets[qm.from() - 1].playerId()) + " bullet=" + bullets.size());
                    }
                }
            }
        }
    }

    private void updateBulletsPositions(long deltaTime) {
        double deltaMs = deltaTime / Spec.MICROS_PER_MS;
        for (Bullet bullet : bullets) {
            bullet.rect.x += deltaMs * bullet.velocity[0];
            bullet.rect.y += deltaMs * bullet.velocity[1];
        }
    }

    private void checkBulletBulletCollisions() {
        for (int i = 0; i < bullets.size(); i++) {
            for (int j = i + 1; j < bullets.size(); j++) {
                if (bullets.get(i).rect.collides(bullets.get(j).rect)) {
                    bullets.remove(j);
                    bullets.remove(i);
                    return;
                }
            }
        }
    }

    private Player checkBulletPlayerCollisions() {
        for (Player player : players) {
            for (Bullet bullet : bullets) {
                if (bullet.rect.collides(player.rect)) {
                    return player;
                }
            }
        }
        return null;
    }

    private Socket getPlayerSocket(Player player) {
        if (player == players[0]) return sockets[0];
        return sockets[1];
    }

    private Player getOtherPlayer(Player player) {
        if (player == players[0]) return players[1];
        return players[0];
    }

    public void run() {
        start();
        ActiveGames.add();

        try {
            Logger log = Logger.forPair(sockets[0].playerId(), sockets[1].playerId());

            sockets[0].send(new Message(Message.GAME_ON));
            sockets[1].send(new Message(Message.GAME_ON));

            log.logInfo("game on", "");

            long winnerId = 0;
            int ticks = 0;
            Instant tickStartTime = Instant.now();

            long lastLoopTime = System.nanoTime() / 1000;

            while (true) {
                ticks++;

                long startTime = System.nanoTime() / 1000;
                long deltaTime = startTime - lastLoopTime;

                if (ticks > 1) {
                    stats.addDeltaTime(deltaTime);
                }

                updateStateFromMessageQueue(log);
                updateBulletsPositions(deltaTime);
                checkBulletBulletCollisions();

                Player loser = checkBulletPlayerCollisions();
                if (loser != null) {
                    Player winner = getOtherPlayer(loser);
                    Socket winnerSocket = getPlayerSocket(winner);
                    Socket loserSocket = getPlayerSocket(loser);
                    winnerId = winnerSocket.playerId();

                    winnerSocket.send(Message.createWinner(winnerId, stats.toString(), ActiveGames.get()));
                    loserSocket.send(Message.createLoser());

                    winnerSocket.close();
                    loserSocket.close();

                    break;
                }

                long nowTime = System.nanoTime() / 1000;
                long sleepUs = Spec.TICK_TARGET_MICROS - (nowTime - startTime);
                if (sleepUs > 0) {
                    Thread.sleep(Duration.ofNanos(sleepUs * 1000));
                }
                lastLoopTime = startTime;
            }

            Duration elapsed = Duration.between(tickStartTime, Instant.now());
            String elapsedStr = String.format("%.4fs", elapsed.toNanos() / 1_000_000_000.0);
            log.logMilestone("game over", String.format(
                "winner=%d histogram=%s active_games=%d ticks=%d elapsed=%s bullets=%d",
                winnerId,
                stats,
                ActiveGames.get(),
                ticks,
                elapsedStr,
                bullets.size()
            ));
        } catch (InterruptedException _) {
        } finally {
            stop();
            ActiveGames.remove();
        }
    }
}
