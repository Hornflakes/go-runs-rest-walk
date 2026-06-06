package com.hornflakes.gorunsrestwalk.gameloop;

import com.hornflakes.gorunsrestwalk.logger.Logger;
import com.hornflakes.gorunsrestwalk.server.Message;
import com.hornflakes.gorunsrestwalk.server.Socket;
import com.hornflakes.gorunsrestwalk.stats.ActiveGames;
import com.hornflakes.gorunsrestwalk.stats.GameFrameStats;

import java.util.ArrayList;
import java.util.concurrent.locks.LockSupport;

import static com.hornflakes.gorunsrestwalk.gameloop.Spec.*;

public final class Game {

    private final Socket[] sockets;
    private final Player[] players;
    private final ArrayList<Bullet> bullets = new ArrayList<>();
    private final Queue queue = new Queue();
    private final GameFrameStats stats = new GameFrameStats();
    private final boolean verbose;

    public Game(Socket s0, Socket s1, boolean verbose) {
        this.sockets = new Socket[]{s0, s1};
        this.players = new Player[]{
                new Player(PLAYER0_SPAWN, PLAYER0_DIR, PLAYER0_FIRE_RATE_MS),
                new Player(PLAYER1_SPAWN, PLAYER1_DIR, PLAYER1_FIRE_RATE_MS),
        };
        this.verbose = verbose;
    }

    public void run() {
        queue.start(sockets[0], sockets[1]);
        ActiveGames.increment();

        try {
            gameLoop();
        } finally {
            queue.stop();
            ActiveGames.decrement();
        }
    }

    private void gameLoop() {
        var log = Logger.forPair(sockets[0].playerId(), sockets[1].playerId());

        sockets[0].send(Message.create(Message.GAME_ON));
        sockets[1].send(Message.create(Message.GAME_ON));
        log.info("game on", "");

        long winnerId = 0;
        int ticks = 0;
        long tickStartNanos = System.nanoTime();
        long lastLoopMicros = System.nanoTime() / 1000;

        for (;;) {
            ticks++;

            long startMicros = System.nanoTime() / 1000;
            long deltaTime = startMicros - lastLoopMicros;

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

            long nowMicros = System.nanoTime() / 1000;
            long sleepUs = TICK_TARGET_MICROS - (nowMicros - startMicros);
            if (sleepUs > 0) {
                LockSupport.parkNanos(sleepUs * 1000);
            }
            lastLoopMicros = startMicros;
        }

        long elapsedNanos = System.nanoTime() - tickStartNanos;
        String elapsed = String.format("%.7fs", elapsedNanos / 1_000_000_000.0);
        log.milestone("game over",
                "winner=" + winnerId
                        + " histogram=" + stats
                        + " active_games=" + ActiveGames.get()
                        + " ticks=" + ticks
                        + " elapsed=" + elapsed
                        + " bullets=" + bullets.size());
    }

    private void updateStateFromMessageQueue(Logger log) {
        var messages = queue.flush();
        if (messages == null) return;

        for (var qm : messages) {
            if (qm.message().getType() == Message.SHOOT) {
                Player player = players[qm.from() - 1];
                if (player.fire()) {
                    bullets.add(Bullet.fromPlayer(player));

                    if (verbose) {
                        log.info("player shot",
                                Logger.player(sockets[qm.from() - 1].playerId())
                                        + " bullet=" + bullets.size());
                    }
                }
            }
        }
    }

    private void updateBulletsPositions(long deltaTimeMicros) {
        double deltaMs = deltaTimeMicros / MICROS_PER_MS;
        for (int i = 0, n = bullets.size(); i < n; i++) {
            bullets.get(i).updatePosition(deltaMs);
        }
    }

    private void checkBulletBulletCollisions() {
        int size = bullets.size();
        for (int i = 0; i < size; i++) {
            for (int j = i + 1; j < size; j++) {
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
            for (int i = 0, n = bullets.size(); i < n; i++) {
                if (bullets.get(i).rect.collides(player.rect)) {
                    return player;
                }
            }
        }
        return null;
    }

    private Socket getPlayerSocket(Player player) {
        return player == players[0] ? sockets[0] : sockets[1];
    }

    private Player getOtherPlayer(Player player) {
        return player == players[0] ? players[1] : players[0];
    }
}
