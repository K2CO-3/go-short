#!/bin/bash

# 实时监控延迟统计（配合 wrk 压测使用）
# 使用方法: ./scripts/monitor.sh [刷新间隔秒数]

INTERVAL=${1:-1}

echo "=========================================="
echo "📊 延迟监控 (每 ${INTERVAL}s 刷新) - 按 Ctrl+C 退出"
echo "=========================================="
echo ""

while true; do
    clear
    echo "=========================================="
    echo "📊 $(date '+%Y-%m-%d %H:%M:%S')"
    echo "=========================================="
    echo ""
    curl -s http://localhost/metrics 2>/dev/null | jq '.' 2>/dev/null || curl -s http://localhost/metrics
    echo ""
    echo "按 Ctrl+C 退出"
    sleep $INTERVAL
done
