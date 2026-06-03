import { WebSocket } from 'ws';
import { Message, unmarshalMessage } from './message.js';
import * as logger from '../logger/index.js';

export type Socket = {
    remoteAddr: string;
    playerId: number;
    disconnected: boolean;
    closed: boolean;
    onMessage: ((msg: Message) => void) | null;
    send(msg: Message): void;
    close(): void;
};

function normalReadEnd(code: number, err: unknown): boolean {
    if (code === 1000 || code === 1001) {
        return true;
    }
    if (err instanceof Error && err.message.includes('use of closed network connection')) {
        return true;
    }
    return false;
}

export function createSocket(ws: WebSocket, remoteAddr: string): Socket {
    let readEnded = false;

    const socket: Socket = {
        remoteAddr,
        playerId: 0,
        disconnected: false,
        closed: false,
        onMessage: null,

        send(msg: Message): void {
            if (socket.disconnected || socket.closed) return;

            let data: string;
            try {
                data = JSON.stringify(msg);
            } catch (err) {
                logger.softError('websocket message marshal failed', logDetail(err));
                return;
            }

            ws.send(data, (err) => {
                if (!err) return;
                if (socket.disconnected || socket.closed) return;

                logger.hardError('websocket message write failed', logDetail(err));
            });
        },

        close(): void {
            if (socket.closed) return;

            socket.closed = true;
            ws.close();
        },
    };

    function logDetail(err: unknown): string {
        return `${logger.playerWithAddr(socket.playerId, socket.remoteAddr)} err=${err}`;
    }

    function markReadEnded(err: unknown, code?: number): void {
        if (readEnded) return;

        readEnded = true;

        const codeVal = code ?? 0;
        if (!normalReadEnd(codeVal, err)) {
            logger.warn('websocket read ended', logDetail(err));
        }

        socket.disconnected = true;
    }

    ws.on('message', (data, isBinary) => {
        if (isBinary) return;

        try {
            const msg = unmarshalMessage(data.toString());
            if (socket.onMessage) {
                socket.onMessage(msg);
            }
        } catch (err) {
            logger.softError('websocket message unmarshal failed', logDetail(err));
        }
    });

    ws.on('close', (code, reason) => {
        const err = reason.length > 0 ? reason.toString() : code;
        markReadEnded(err, code);
    });

    ws.on('error', (err) => {
        markReadEnded(err);
    });

    return socket;
}
