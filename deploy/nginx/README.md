# 网关 Nginx（HTTPS）

`docker-compose.yml` 中的 `nginx` 服务使用本目录配置：

- `nginx.conf`：监听 **80**（跳转 HTTPS）与 **443**（TLS），反向代理到 `frontend-service`、`api-service`、`redirect-service`。
- `certs/`：`server.crt` / `server.key` 供 TLS 使用（见下方）。

## 证书

仓库内附带 **开发用自签名** 证书（`CN=localhost`，含 SAN）。首次部署若缺失证书，可在本目录执行：

```bash
chmod +x gen-cert.sh && ./gen-cert.sh
```

生产环境请替换为 **Let's Encrypt** 或企业 CA 签发的证书，保持文件名仍为 `certs/server.crt` 与 `certs/server.key`，或修改 `nginx.conf` 中 `ssl_certificate` 路径后 `docker compose up -d --force-recreate nginx`。

## 访问

- `https://localhost/`（或本机 IP）访问前端；HTTP 80 会 **301** 到 HTTPS。
- 自签名证书下浏览器可能告警，开发环境可选择继续访问。
