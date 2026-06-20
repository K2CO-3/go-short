# GoShort Frontend

基于 `openapi.yaml` 实现的前端控制台（TypeScript + React + Vite），覆盖当前后端全部 `/api/v1` 接口：

- 认证：注册、登录
- 用户：资料查看/更新、修改密码
- 链接：创建、列表、删除、按别名查询
- 管理员：用户管理、访问日志筛选

## 运行

```bash
cd frontend
npm install
npm run dev
```

默认请求地址：

- 默认使用 `/api/v1`，开发态由 Vite 代理到 `http://localhost:8080`
- 也可通过 `.env` 覆盖（例如你走 Nginx HTTPS 网关）：

```bash
VITE_API_BASE_URL=https://localhost/api/v1
```

## 端口说明

- 前端开发服务器：`http://localhost:5173`
- 后端 API 服务：`http://localhost:8080`
- 当前已配置 Vite 代理，浏览器请求 `5173/api/v1/...` 会自动转发到 `8080/api/v1/...`

## Nginx 部署模式（推荐外网/内网穿透）

项目已支持把前端构建为静态文件并通过 **网关 Nginx**（`deploy/nginx`）统一暴露 **HTTPS**：

- 网关监听 **443**（TLS 证书见 `deploy/nginx/certs/`），**80** 会重定向到 HTTPS
- 前端容器内仍为 HTTP:80，由网关反向代理；镜像内配置见 `frontend/nginx/default.conf`
- `/` → 前端静态站点；`/api/` → API；`/code/` → 跳转服务

启动：

```bash
cd deploy
# 若尚无证书：bash nginx/gen-cert.sh
docker compose up -d --build frontend-service nginx api-service redirect-service
```

访问（请使用 HTTPS）：

- `https://localhost/` 或 `https://<你的域名或IP>/`
- 开发用自签名证书时浏览器可能提示风险，可手动继续访问
