import { createServer } from 'http';
import { WebSocketServer } from 'ws';
import * as logger from './logger.js';

const server = createServer();
const wss = new WebSocketServer({ server, path: '/' });

wss.on('connection', (ws, req) => {
    const ip = req.socket.remoteAddress?.replace('::ffff:', '') ?? '';
    const addr = ip + ':' + req.socket.remotePort;
    logger.info('websocket connected', `addr=${addr}`);
});

server.listen(37373, () => {
    logger.info('server listening', 'addr=:37373');
});
