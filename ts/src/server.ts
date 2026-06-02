import { WebSocket } from 'ws';
import { Socket, createSocket } from './socket.js';
import { createHelloMessage } from './message.js';
import * as logger from './logger.js';
import { IncomingMessage } from 'http';

export type Pair = [Socket, Socket];

export interface Server {
    onPair: ((pair: Pair) => void) | null;
    handleConnection(ws: WebSocket, req: IncomingMessage): void;
}

function socketAlive(s: Socket | null): boolean {
    return s !== null && !s.closed && !s.disconnected;
}

export function createServer(): Server {
    let nextPlayerId = 0;
    let waitingSocket: Socket | null = null;

    const srv: Server = {
        onPair: null,

        handleConnection(ws: WebSocket, req: IncomingMessage): void {
            const ip = req.socket.remoteAddress?.replace('::ffff:', '') ?? '';
            const addr = ip + ':' + req.socket.remotePort;
            const socket = createSocket(ws, addr);

            nextPlayerId++;
            socket.playerId = nextPlayerId;

            socket.send(createHelloMessage(socket.playerId));
            logger.info('websocket connected', logger.playerWithAddr(socket.playerId, addr));

            if (socketAlive(waitingSocket)) {
                const pair: Pair = [waitingSocket!, socket];
                waitingSocket = null;

                if (srv.onPair) {
                    srv.onPair(pair);
                }
                return;
            }

            if (waitingSocket !== null) {
                waitingSocket.close();
                waitingSocket = null;
            }

            waitingSocket = socket;
        },
    };

    return srv;
}
