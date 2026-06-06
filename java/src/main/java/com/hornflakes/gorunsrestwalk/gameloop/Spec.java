package com.hornflakes.gorunsrestwalk.gameloop;

public final class Spec {

    private Spec() {}

    public static final double PLAYER_WIDTH = 100;
    public static final double PLAYER_HEIGHT = 100;
    public static final double BULLET_WIDTH = 35;
    public static final double BULLET_HEIGHT = 3;

    public static final double PLAYER0_SPAWN_X = 2500;
    public static final double PLAYER1_SPAWN_X = -2500;
    public static final long PLAYER0_FIRE_RATE_MS = 180;
    public static final long PLAYER1_FIRE_RATE_MS = 300;

    public static final double[] PLAYER0_SPAWN = {PLAYER0_SPAWN_X, 0};
    public static final double[] PLAYER1_SPAWN = {PLAYER1_SPAWN_X, 0};
    public static final double[] PLAYER0_DIR = {-1, 0};
    public static final double[] PLAYER1_DIR = {1, 0};

    public static final double BULLET_SPEED_MS = 1.0;
    public static final long TICK_TARGET_MICROS = 16_000;
    public static final double MICROS_PER_MS = 1000;

    public static final long READY_TIMEOUT_MS = 30_000;
}
