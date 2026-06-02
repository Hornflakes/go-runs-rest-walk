import { createServer as createHttpServer } from 'http';
import { WebSocketServer } from 'ws';
import * as logger from './logger.js';
import { createServer } from './server.js';
import { Logger } from './logger.js';

const httpServer = createHttpServer();
const wss = new WebSocketServer({ server: httpServer, path: '/' });

const srv = createServer();

srv.onPair = (pair) => {
    const [s0, s1] = pair;
    const log = new Logger(s0.playerId, s1.playerId);
    log.milestone('websockets paired', '');
};

wss.on('connection', (ws, req) => {
    srv.handleConnection(ws, req);
});

httpServer.listen(37373, () => {
    logger.info('server listening', 'addr=:37373');
});
