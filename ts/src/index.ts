import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import * as logger from './logger.js';
import { createSocket } from './socket.js';

const server = createServer();
const wss = new WebSocketServer({ server, path: '/' });

wss.on('connection', (ws, req) => {
    const ip = req.socket.remoteAddress?.replace('::ffff:', '') ?? '';
    const addr = ip + ':' + req.socket.remotePort;

    const socket = createSocket(ws, addr);
    logger.info('websocket connected', `addr=${addr}`);

    socket.onMessage = (msg) => {
        logger.info(
            'echo',
            `${logger.playerWithAddr(socket.playerId, addr)} msg=${JSON.stringify(msg)}`,
        );
        socket.send(msg);
    };
});

server.listen(37373, () => {
    logger.info('server listening', 'addr=:37373');
});
