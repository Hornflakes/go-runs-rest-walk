export type Vector2D = [number, number];

export class Rect {
    x: number;
    y: number;
    width: number;
    height: number;

    constructor(x: number, y: number, width: number, height: number) {
        this.x = x;
        this.y = y;
        this.width = width;
        this.height = height;
    }

    setPosition(x: number, y: number): void {
        this.x = x;
        this.y = y;
    }

    collides(other: Rect): boolean {
        if (this.x > other.x + other.width || other.x > this.x + this.width) {
            return false;
        }
        if (this.y > other.y + other.height || other.y > this.y + this.height) {
            return false;
        }
        return true;
    }
}
