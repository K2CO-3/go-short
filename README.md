# GoShort

一个自研的短链接服务，用来练手和学习。

---

## 声明

这是一个**个人学习项目**，不是生产级产品。写它的动机很简单：想亲手搭一套「短链接」系统，从零体验高并发场景下的缓存设计、异步队列、布隆过滤器这些常被挂在嘴边的东西。代码里有很多试错痕迹，也有不少可以优化的地方，但正是这些过程让人对「为什么这样设计」有了更实诚的理解。

如果你也在学 Go 或后端架构，欢迎参考或一起琢磨；如果你发现了 bug 或更好的实现方式，也欢迎提 issue 或 PR。

---

## 快速启动

```bash
cd deploy
docker-compose up --build
```

启动后（默认 compose 端口）：

- **Nginx 入口**：`http://localhost`（或按 `BASE_URL` / 证书使用 HTTPS）
- **跳转服务（直连调试用）**：`http://localhost:8082`（如 `/code/{短码}`）
- **API（直连调试用）**：`http://localhost:8081`，路径前缀一般为 `/api/v1`
- 健康检查：各服务常见路径为 `GET /health`（以对应 `main` 与路由为准）

**前端**：静态资源由 **frontend 镜像** 构建，经 Nginx 与 API、跳转同域或反代。本地改界面可在/frontend目录下：

```bash
cd frontend
npm install   # 或 npm ci（见下方 package-lock.json 说明）
npm run dev
```

更完整的架构、端口与 API 列表见 [短链接服务文档.md](短链接服务文档.md)。

---

## 项目结构

| 部分 | 说明 |
|------|------|
| **cmd/redirect-server** | 跳转服务；本地缓存 + Redis + 布隆；访问日志写入 **Kafka** |
| **cmd/api-server** | 管理 API；用户、短链、资料、**管理员**接口 |
| **cmd/worker** | 从 **Kafka** 消费访问日志，写入 PostgreSQL |
| **frontend/** | Vite + React + TypeScript 管理端 |

---
