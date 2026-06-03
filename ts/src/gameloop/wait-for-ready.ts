import { Socket } from '../server/socket.js';
import { createReadyMessage, MessageType } from '../server/message.js';
import { DisconnectPollMs, ReadyTimeoutMs } from './spec.js';

export type ReadyHandle = {
    promise: Promise<boolean>;
    cancel: () => void;
};

export function waitForReady(s0: Socket, s1: Socket): ReadyHandle {
    let cancelled = false;
    let disconnectInterval: ReturnType<typeof setInterval> | null = null;
    let timeoutHandle: ReturnType<typeof setTimeout> | null = null;
    let resolve: (ok: boolean) => void;

    const promise = new Promise<boolean>((res) => {
        resolve = res;
    });

    function cleanup(): void {
        cancelled = true;
        if (disconnectInterval !== null) {
            clearInterval(disconnectInterval);
            disconnectInterval = null;
        }
        if (timeoutHandle !== null) {
            clearTimeout(timeoutHandle);
            timeoutHandle = null;
        }
        s0.onMessage = null;
        s1.onMessage = null;
    }

    s0.send(createReadyMessage(s1.playerId));
    s1.send(createReadyMessage(s0.playerId));

    let count = 0;

    s0.onMessage = (msg) => {
        if (cancelled) return;
        if (msg.type === MessageType.Ready) {
            count++;
            s0.onMessage = null;
            if (count === 2) {
                cleanup();
                resolve(true);
            }
        }
    };

    s1.onMessage = (msg) => {
        if (cancelled) return;
        if (msg.type === MessageType.Ready) {
            count++;
            s1.onMessage = null;
            if (count === 2) {
                cleanup();
                resolve(true);
            }
        }
    };

    disconnectInterval = setInterval(() => {
        if (s0.disconnected || s0.closed || s1.disconnected || s1.closed) {
            cleanup();
            resolve(false);
        }
    }, DisconnectPollMs);

    timeoutHandle = setTimeout(() => {
        if (!cancelled) {
            cleanup();
            resolve(false);
        }
    }, ReadyTimeoutMs);

    function cancel(): void {
        if (!cancelled) {
            cleanup();
            resolve(false);
        }
    }

    return { promise, cancel };
}
