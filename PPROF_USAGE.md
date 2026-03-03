# pprof 使用说明

Go 标准库 pprof，**零埋点**：`import _ "net/http/pprof"` 即生效，无需在业务代码里加 profiling 逻辑。

## 用法

redirect-server 已接入，启动后访问：

```
http://localhost:6060/debug/pprof/
```

本地直连进程；若用 Docker，需暴露 6060 端口。

## 常用命令

```bash
# CPU（采样 30 秒）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

进入后可用 `top`、`list 函数名`、`web`（需 graphviz）等。

## 与 metrics 的区别

| 项目   | pprof | metrics (internal/metrics)     |
|--------|-------|--------------------------------|
| 作用   | 运行时代码级分析（CPU、内存、goroutine） | 业务层延迟与命中率（本地缓存 / Redis / DB） |
| 埋点   | 无，标准库自动采集 | 需在 handler 中 `RecordXxx()` 埋点 |
| 访问   | `:6060/debug/pprof/` | `/metrics`（经 Nginx 为 `http://localhost/metrics`） |
| 场景   | 排查 CPU 高、内存泄漏、goroutine 堆积 | 观察缓存效果和各层延迟 |
