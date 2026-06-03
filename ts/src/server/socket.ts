import { WebSocket } from 'ws';
import { Message, unmarshalMessage } from './message.js';
import * as logger from '../logger/index.js';

export interface Socket {
    remoteAddr: string;
    playerId: number;
    disconnected: boolean;
    closed: boolean;
    onMessage: ((msg: Message) => void) | null;
    send(msg: Message): void;
    close(): void;
}

export function createSocket(ws: WebSocket, remoteAddr: string): Socket {
    const socket: Socket = {
        remoteAddr,
        playerId: 0,
        disconnected: false,
        closed: false,
        onMessage: null,

        send(msg: Message): void {
            if (socket.disconnected || socket.closed) return;
            const data = JSON.stringify(msg);
            ws.send(data);
        },

        close(): void {
            if (socket.closed) return;
            socket.closed = true;
            ws.close();
        },
    };
    const logDetail = `${logger.playerWithAddr(socket.playerId, socket.remoteAddr)}`;

    ws.on('message', (data, isBinary) => {
        if (isBinary) return;

        try {
            const msg = unmarshalMessage(data.toString());
            if (socket.onMessage) {
                socket.onMessage(msg);
            }
        } catch (err) {
            logger.softError('websocket message unmarshal failed', `${logDetail} err=${err}`);
        }
    });

    ws.on('close', () => {
        socket.disconnected = true;
    });

    ws.on('error', (err) => {
        if (socket.disconnected) return;
        logger.hardError('websocket error', `${logDetail} err=${err}`);
        socket.disconnected = true;
    });

    return socket;
}
