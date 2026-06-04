package com.hornflakes.gorunsrestwalk.server;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

@JsonInclude(JsonInclude.Include.NON_EMPTY)
public record Message(
    @JsonProperty("type") int type,
    @JsonProperty("msg") String msg
) {
    public static final int HELLO = 0;
    public static final int READY = 1;
    public static final int GAME_ON = 2;
    public static final int SHOOT = 3;
    public static final int GAME_OVER = 4;

    private static final ObjectMapper MAPPER = new ObjectMapper();

    public Message(int type) {
        this(type, "");
    }

    public static Message unmarshal(String json) throws JsonProcessingException {
        return MAPPER.readValue(json, Message.class);
    }

    public String marshal() throws JsonProcessingException {
        return MAPPER.writeValueAsString(this);
    }

    public static Message createHello(long playerId) {
        return new Message(HELLO, "playerId=" + playerId);
    }

    public static Message createReady(long enemyId) {
        return new Message(READY, "enemyId=" + enemyId);
    }

    public static Message createWinner(long winnerId, String histogram, long activeGames) {
        return new Message(GAME_OVER, "winner=" + winnerId + " histogram=" + histogram + " active_games=" + activeGames);
    }

    public static Message createLoser() {
        return new Message(GAME_OVER, "loser");
    }

    public static long parseHelloId(String msg) {
        return Long.parseUnsignedLong(msg.substring("playerId=".length()));
    }

    public static long parseReadyId(String msg) {
        return Long.parseUnsignedLong(msg.substring("enemyId=".length()));
    }
}
