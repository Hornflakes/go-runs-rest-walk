export enum MessageType {
    Hello,
    Ready,
    GameOn,
    Shoot,
    GameOver,
}

export interface Message {
    type: MessageType;
    msg?: string;
}

export function unmarshalMessage(data: string): Message {
    return JSON.parse(data) as Message;
}

export function createMessage(msgType: number): Message {
    return { type: msgType };
}

export function createHelloMessage(playerId: number): Message {
    return { type: MessageType.Hello, msg: `playerId=${playerId}` };
}

export function createReadyMessage(enemyId: number): Message {
    return { type: MessageType.Ready, msg: `enemyId=${enemyId}` };
}

export function createWinnerMessage(
    winnerId: number,
    histogram: string,
    activeGames: number,
): Message {
    return {
        type: MessageType.GameOver,
        msg: `winner=${winnerId} histogram=${histogram} active_games=${activeGames}`,
    };
}

export function createLoserMessage(): Message {
    return { type: MessageType.GameOver, msg: 'loser' };
}

export function parseHelloMessage(msg: string): number {
    return parseInt(msg.replace('playerId=', ''), 10);
}

export function parseReadyMessage(msg: string): number {
    return parseInt(msg.replace('enemyId=', ''), 10);
}
