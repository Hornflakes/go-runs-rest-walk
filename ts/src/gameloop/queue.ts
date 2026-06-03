import { Socket } from '../server/socket.js';
import { Message } from '../server/message.js';

export interface QueueMessage {
    from: number;
    message: Message;
}

export class Queue {
    private messages: QueueMessage[] = [];
    private stopped = false;

    start(s0: Socket, s1: Socket): void {
        s0.onMessage = (msg) => {
            if (this.stopped) return;
            this.messages.push({ from: 1, message: msg });
        };

        s1.onMessage = (msg) => {
            if (this.stopped) return;
            this.messages.push({ from: 2, message: msg });
        };
    }

    stop(): void {
        this.stopped = true;
    }

    flush(): QueueMessage[] | null {
        if (this.messages.length === 0) {
            return null;
        }
        const messages = this.messages;
        this.messages = [];
        return messages;
    }
}
