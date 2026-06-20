#!/bin/bash

# 生成自签名 SSL 证书（仅用于开发/测试）
# 生产环境请使用 Let's Encrypt 或正规 CA 证书

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CERT_DIR="$SCRIPT_DIR/certs"
mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout server.key \
  -out server.crt \
  -config "$SCRIPT_DIR/openssl-local.conf" \
  -extensions v3_req

chmod 600 server.key
echo "✅ 证书已生成: $CERT_DIR/server.crt"
echo "   私钥: $CERT_DIR/server.key"
echo "   有效期: 365 天（含 localhost / 127.0.0.1 SAN）"
echo ""
echo "⚠️  自签名证书仅供开发使用；浏览器可能仍提示不受信任，可添加安全例外。"
