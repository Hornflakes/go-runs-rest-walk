#!/usr/bin/env bash
set -euo pipefail

pid=$1
out=$2
interval=${3:-1}

if ! kill -0 "$pid" 2>/dev/null; then
  echo "pid $pid not running" >&2
  exit 1
fi

echo "Type,Value" > "$out"

while kill -0 "$pid" 2>/dev/null; do
  line=$(pidstat -p "$pid" "$interval" 1 2>/dev/null | tail -1)
  cpu=$(echo "$line" | awk '{print $8}')

  rline=$(pidstat -r -p "$pid" "$interval" 1 2>/dev/null | tail -1)
  rss_kb=$(echo "$rline" | awk '{print $7}')
  mem_mb=$(awk -v kb="$rss_kb" 'BEGIN { printf "%.2f MB", kb/1024 }')

  echo "CPU,$cpu" >> "$out"
  echo "MEM,$mem_mb" >> "$out"
done
