#!/bin/bash

# 压测脚本：使用 wrk 测试重定向服务性能
#
# 使用方法:
#   ./scripts/benchmark.sh [short_code] [duration] [concurrent]
#   ./scripts/benchmark.sh test123 30 500     # 30秒，500并发
#   ./scripts/benchmark.sh test123 0 500      # 持续压测，直到 Ctrl+C
#
# 参数说明:
#   short_code: 短链接代码，如 test123
#   duration:   压测时长(秒)，0 表示持续压测直到 Ctrl+C
#   concurrent: 并发数 = 同时保持的 HTTP 连接数（不是 QPS）

set -e

SHORT_CODE=${1:-"test123"}
DURATION=${2:-30}
CONCURRENT=${3:-500}

echo "=========================================="
echo "🚀 重定向服务性能压测 (wrk)"
echo "=========================================="
echo "短链接代码: $SHORT_CODE"
if [ "$DURATION" -eq 0 ]; then
    echo "压测模式: 持续压测（按 Ctrl+C 停止）"
else
    echo "压测时长: ${DURATION}秒"
fi
echo "并发数: $CONCURRENT"
echo "=========================================="
echo ""

# 检查 wrk 是否安装
if ! command -v wrk &> /dev/null; then
    echo "❌ 错误: wrk 未安装"
    echo "   Ubuntu/Debian: sudo apt-get install wrk"
    echo "   macOS: brew install wrk"
    exit 1
fi

# 检查服务是否运行（经 Nginx 80 端口）
if ! curl -s http://localhost/health > /dev/null 2>&1; then
    echo "❌ 错误: 服务未运行，请先启动（Nginx + redirect-server）"
    exit 1
fi

echo "📊 压测前延迟统计:"
echo "----------------------------------------"
curl -s http://localhost/metrics | jq '.' 2>/dev/null || curl -s http://localhost/metrics
echo ""
echo ""

echo "🔥 开始压测..."
echo ""

URL="http://localhost/code/$SHORT_CODE"
if [ "$DURATION" -eq 0 ]; then
    wrk -t4 -c$CONCURRENT -d86400s --latency "$URL"
else
    wrk -t4 -c$CONCURRENT -d${DURATION}s --latency "$URL"
fi

echo ""
echo "=========================================="
echo "📊 压测后延迟统计:"
echo "----------------------------------------"
curl -s http://localhost/metrics | jq '.' 2>/dev/null || curl -s http://localhost/metrics
echo ""
echo ""

echo "💡 提示: 实时监控 ./scripts/monitor.sh"
