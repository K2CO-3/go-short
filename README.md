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

启动后访问：
- 短链跳转：`http://localhost/` 或 `http://localhost:8082/`（视 Nginx 配置）
- API 服务：`http://localhost:8081/`
- 健康检查：`http://localhost:8081/health`、`http://localhost:8082/health`

---

## 项目结构

三个独立服务：

| 服务 | 说明 |
|------|------|
| **redirect-server** | 跳转服务，本地缓存 + Redis + 布隆过滤器，处理读流量 |
| **api-server** | 管理 API，用户、链接、管理员功能 |
| **worker** | 消费 Redis Stream，异步写入访问日志 |

更多细节见 [短链接服务文档.md](短链接服务文档.md)。
