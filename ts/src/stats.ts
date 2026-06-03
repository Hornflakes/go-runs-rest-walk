export let activeGames = 0;

export function addActiveGame(): void {
    activeGames++;
}

export function removeActiveGame(): void {
    activeGames--;
}

export class GameFrameStats {
    frameBuckets: number[] = [0, 0, 0, 0, 0, 0, 0, 0];

    addDeltaTime(delta: number): void {
        if (delta > 40_999) {
            this.frameBuckets[7]++;
        } else if (delta > 30_999) {
            this.frameBuckets[6]++;
        } else if (delta > 25_999) {
            this.frameBuckets[5]++;
        } else if (delta > 23_999) {
            this.frameBuckets[4]++;
        } else if (delta > 21_999) {
            this.frameBuckets[3]++;
        } else if (delta > 19_999) {
            this.frameBuckets[2]++;
        } else if (delta > 17_999) {
            this.frameBuckets[1]++;
        } else {
            this.frameBuckets[0]++;
        }
    }

    toString(): string {
        return this.frameBuckets.join(',');
    }
}
