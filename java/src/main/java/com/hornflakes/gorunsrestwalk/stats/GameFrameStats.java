package com.hornflakes.gorunsrestwalk.stats;

import java.util.StringJoiner;

public class GameFrameStats {

    private final long[] frameBuckets = new long[8];

    public void addDeltaTime(long delta) {
        if (delta > 40_999) {
            frameBuckets[7]++;
        } else if (delta > 30_999) {
            frameBuckets[6]++;
        } else if (delta > 25_999) {
            frameBuckets[5]++;
        } else if (delta > 23_999) {
            frameBuckets[4]++;
        } else if (delta > 21_999) {
            frameBuckets[3]++;
        } else if (delta > 19_999) {
            frameBuckets[2]++;
        } else if (delta > 17_999) {
            frameBuckets[1]++;
        } else {
            frameBuckets[0]++;
        }
    }

    @Override
    public String toString() {
        StringJoiner sj = new StringJoiner(",");
        for (long bucket : frameBuckets) {
            sj.add(Long.toString(bucket));
        }
        return sj.toString();
    }
}
