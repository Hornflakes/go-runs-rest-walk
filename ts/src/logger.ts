enum Level {
    Milestone,
    Info,
    Warn,
    SoftError,
    HardError,
}

const COLORS: Record<Level, [string, string]> = {
    [Level.Milestone]: ['\x1b[32m', '\x1b[0m'], // green
    [Level.Info]: ['\x1b[36m', '\x1b[0m'], // cyan
    [Level.Warn]: ['\x1b[33m', '\x1b[0m'], // yellow
    [Level.SoftError]: ['\x1b[35m', '\x1b[0m'], // magenta
    [Level.HardError]: ['\x1b[31m', '\x1b[0m'], // red
};

function colorize(level: Level, text: string): string {
    const [open, close] = COLORS[level];
    return open + text + close;
}

function formatPrefix(prefix: string): string {
    if (prefix === '') return '';
    return `[${prefix}] `;
}

function write(prefix: string, level: Level, event: string, detail: string): void {
    const p = formatPrefix(prefix);
    const e = colorize(level, event);
    const now = new Date();
    const date = now.toISOString().slice(0, 10).replace(/-/g, '/');
    const time = now.toLocaleTimeString('en-GB', { hour12: false });
    const timestamp = `${date} ${time}`;

    if (detail === '') {
        process.stdout.write(`${timestamp} ${p}${e}\n`);
        return;
    }
    process.stdout.write(`${timestamp} ${p}${e} | ${detail}\n`);
}

export function milestone(event: string, detail: string): void {
    write('', Level.Milestone, event, detail);
}
export function info(event: string, detail: string): void {
    write('', Level.Info, event, detail);
}
export function warn(event: string, detail: string): void {
    write('', Level.Warn, event, detail);
}
export function softError(event: string, detail: string): void {
    write('', Level.SoftError, event, detail);
}
export function hardError(event: string, detail: string): void {
    write('', Level.HardError, event, detail);
}

export function player(id: number): string {
    return `player=${id}`;
}

export function playerWithAddr(id: number, addr: string): string {
    return `player=${id} addr=${addr}`;
}

export function playerPair(playerId0: number, playerId1: number): string {
    return `${player(playerId0)} vs ${player(playerId1)}`;
}

export function pairPrefix(playerId0: number, playerId1: number): string {
    return `${playerId0} vs ${playerId1}`;
}

export class Logger {
    private prefix: string;

    constructor(playerId0: number, playerId1: number) {
        this.prefix = pairPrefix(playerId0, playerId1);
    }

    milestone(event: string, detail: string): void {
        write(this.prefix, Level.Milestone, event, detail);
    }
    info(event: string, detail: string): void {
        write(this.prefix, Level.Info, event, detail);
    }
    warn(event: string, detail: string): void {
        write(this.prefix, Level.Warn, event, detail);
    }
    softError(event: string, detail: string): void {
        write(this.prefix, Level.SoftError, event, detail);
    }
    hardError(event: string, detail: string): void {
        write(this.prefix, Level.HardError, event, detail);
    }
}
