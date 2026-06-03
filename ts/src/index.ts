import { createServer as createHttpServer } from 'http';
import { WebSocketServer } from 'ws';
import * as logger from './logger/index.js';
import { Logger } from './logger/index.js';
import { waitForReady } from './gameloop/wait-for-ready.js';
import { runGame } from './gameloop/index.js';
import { createServer } from './server/index.js';

const verbose = process.argv.includes('--verbose');

const httpServer = createHttpServer();
const wss = new WebSocketServer({ server: httpServer, path: '/' });

const srv = createServer();

srv.onPair = (pair) => {
    const [s0, s1] = pair;
    const log = new Logger(s0.playerId, s1.playerId);
    log.milestone('websockets paired', '');

    const ready = waitForReady(s0, s1);

    ready.promise.then((ok) => {
        if (!ok) {
            log.warn('websockets ready handshake failed', '');

            s0.close();
            s1.close();
            return;
        }

        log.milestone('websockets ready handshake ok', '');

        runGame(s0, s1, verbose);
    });
};

wss.on('connection', (ws, req) => {
    srv.handleConnection(ws, req);
});

wss.on('wsClientError', (err, _socket, req) => {
    const addr = req.socket.remoteAddress ?? '';
    logger.hardError('websocket upgrade failed', `addr=${addr} err=${err}`);
});

httpServer.on('error', (err) => {
    logger.hardError('server listen failed', `addr=:37373 err=${err}`);
    process.exit(1);
});

httpServer.listen(37373, () => {
    logger.info('server listening', 'addr=:37373');
});
