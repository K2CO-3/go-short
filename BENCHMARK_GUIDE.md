# 性能压测和延迟分析指南

压测经 Nginx 接入，流量走 `http://localhost`（80 端口）。

## 🚀 快速开始

### 1. 启动服务

```bash
# Docker：docker compose up -d
# 本地：Nginx + ./redirect-server + PORT=8081 ./api-server
```

### 2. 压测

```bash
./scripts/benchmark.sh test123 30 500
```

---

## 📊 查看延迟统计（Metrics）

`internal/metrics/latency.go` 提供**业务层埋点**：本地缓存 / Redis / PostgreSQL 的命中率和延迟，用于观察缓存效果和各层耗时。与 pprof 不同，需要业务代码调用 `metrics.RecordXxx()`。

```bash
# 经 Nginx（默认 80）
curl http://localhost/metrics

# 或使用 jq 格式化
curl -s http://localhost/metrics | jq

# 实时监控
./scripts/monitor.sh
```

---

## 🔥 完整压测流程

### 步骤1：准备测试数据

确保有一个可用的短链接代码，例如 `test123`。

如果没有，可以通过 API 创建：

```bash
# 1. 登录获取 token
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}' \
  | jq -r '.data.token')

# 2. 创建短链接
curl -X POST http://localhost:8081/api/v1/links \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://www.example.com","short_code":"test123"}'
```

### 步骤2：运行压测脚本

```bash
# 基本用法：./scripts/benchmark.sh <短链接代码> <时长秒> <并发数>
./scripts/benchmark.sh test123 30 500

# 持续压测（直到按 Ctrl+C 停止）：
./scripts/benchmark.sh test123 0 500

# 参数说明：
# - 并发数 = 同时保持的 HTTP 连接数（不是 QPS）
# - 500 并发 ≈ 最多 500 个请求同时进行
# - 实际 QPS = 总请求数 / 耗时，由服务响应速度决定
```

### 步骤3：查看结果

脚本会自动显示：
1. 压测前的延迟统计
2. 压测过程（wrk/ab 输出）
3. 压测后的延迟统计

---

## 📖 并发数是什么意思？

**并发数** = 同时保持的 HTTP 连接数，不是「每秒请求数（QPS）」。

- `-c 500`：最多 500 个连接同时向服务器发请求
- 每个连接发完一个请求会马上发下一个，不等待
- 实际 QPS = 总请求数 ÷ 耗时，由服务响应速度决定
- 并发越高，对服务器的压力越大；过高可能打满连接或带宽

**为什么 ab 很快就结束？**  
ab 按「总请求数」执行，不是按时长。请求数固定，服务器越快，结束得越快。脚本里已改为按预估 QPS 计算请求数，尽量接近指定时长；推荐使用 wrk，原生支持按秒压测。

---

## 🛠️ 手动压测（不使用脚本）

### 使用 wrk（推荐，支持按时长压测）

```bash
# 安装 wrk
# Ubuntu/Debian: sudo apt-get install wrk
# macOS: brew install wrk

# 基本压测（按秒执行，会真正跑满 30 秒）
wrk -t4 -c500 -d30s --latency http://localhost/code/test123

# 持续压测（约 1 小时，可用 Ctrl+C 提前结束）
wrk -t4 -c500 -d3600s --latency http://localhost/code/test123

# 参数说明：
# -t4: 使用4个线程
# -c500: 500个并发连接
# -d30s: 持续30秒（wrk 会真正跑满这么长时间）
# --latency: 显示延迟统计
```

### 使用 Apache Bench (ab)

```bash
# 安装 ab
# Ubuntu/Debian: sudo apt-get install apache2-utils
# macOS: 已内置

# 基本压测（按请求数执行，总请求数发完就结束）
ab -n 100000 -c 500 http://localhost/code/test123

# 参数说明：
# -n 100000: 总共100000个请求
# -c 500: 500个并发
# 注意：ab 不按时长，服务越快结束得越快
```

---

## 📈 多终端压测流程（推荐）

### 终端1：启动服务（Nginx + redirect + api）

### 终端2：实时监控

```bash
./scripts/monitor.sh
```

### 终端3：运行压测

```bash
./scripts/benchmark.sh test123 60 1000
```

### 终端4：pprof（可选）

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=60
```

详见 [PPROF_USAGE.md](./PPROF_USAGE.md)。

---

## 📊 延迟统计解读

### 示例输出

```json
{
  "success": true,
  "data": {
    "local_cache": {
      "hits": 95000,
      "misses": 5000,
      "total": 100000,
      "hit_rate": 95.0,
      "avg_latency_ms": 0.001,
      "max_latency_ms": 0.005
    },
    "redis": {
      "hits": 4500,
      "misses": 500,
      "total": 5000,
      "hit_rate": 90.0,
      "avg_latency_ms": 2.3,
      "max_latency_ms": 8.5
    },
    "postgres": {
      "hits": 500,
      "misses": 0,
      "total": 500,
      "hit_rate": 100.0,
      "avg_latency_ms": 12.5,
      "max_latency_ms": 28.3
    }
  }
}
```

### 指标说明

1. **本地缓存 (local_cache)**
   - `hit_rate`: 命中率，越高越好（理想 > 90%）
   - `avg_latency_ms`: 平均延迟，应该 < 0.01ms
   - 说明：如果命中率低，说明热点数据不够热

2. **Redis**
   - `hit_rate`: 命中率，越高越好（理想 > 80%）
   - `avg_latency_ms`: 平均延迟，应该在 1-5ms
   - 说明：如果延迟 > 10ms，可能是网络问题

3. **PostgreSQL**
   - `hit_rate`: 命中率（查询成功/总数）
   - `avg_latency_ms`: 平均延迟，应该在 5-50ms
   - 说明：如果延迟 > 100ms，需要优化查询或索引

---

## 🎯 性能优化建议

### 如果本地缓存命中率低

1. **增加缓存 TTL**（在 `cmd/redirect-server/main.go` 中）
   ```go
   localCache := local.NewLocalCache(10*time.Minute, 20*time.Minute, 10000)
   ```

2. **预热缓存**：在服务启动时加载热点数据

### 如果 Redis 延迟高

1. **检查网络延迟**
   ```bash
   redis-cli --latency
   ```

2. **优化 Redis 配置**
   - 增加连接池大小
   - 使用 pipeline

### 如果 PostgreSQL 延迟高

1. **检查索引**
   ```sql
   EXPLAIN ANALYZE SELECT * FROM links WHERE short_code = 'test123';
   ```

2. **优化查询**
   - 只查询需要的字段
   - 使用连接池

---

## 📝 完整示例

### 场景：测试重定向服务性能

```bash
# 1. 启动服务（终端1）
./redirect-server

# 2. 实时监控（终端2）
./scripts/monitor.sh

# 3. 运行压测（终端3）
./scripts/benchmark.sh test123 60 1000

# 4. 收集 pprof（终端4，在压测期间）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=60
```

### 预期结果

- **本地缓存命中率**: > 90%
- **Redis 命中率**: > 80%
- **本地缓存延迟**: < 0.01ms
- **Redis 延迟**: 1-5ms
- **PostgreSQL 延迟**: 5-50ms
- **总体 QPS**: > 10,000（取决于硬件）

---

## 🐛 常见问题

### Q: 服务启动失败

```bash
# 检查端口是否被占用
lsof -i :8080
lsof -i :6060

# 或使用其他端口
PORT=8081 ./redirect-server
```

### Q: 压测工具未安装

```bash
# Ubuntu/Debian
sudo apt-get install wrk apache2-utils jq

# macOS
brew install wrk jq
```

### Q: metrics 端点返回空数据

- 先发一些请求预热，再查看：`curl http://localhost/metrics`
- 确认服务正常：`curl http://localhost/health`

### Q: pprof 无法访问

- 6060 为 redirect 进程直连，Docker 需映射该端口
- 本地确保 redirect-server 已启动

---

## 💡 提示

1. **压测前预热**：先发送一些请求预热缓存
2. **逐步增加并发**：从低并发开始，逐步增加
3. **监控系统资源**：使用 `htop` 或 `top` 监控 CPU/内存
4. **对比测试**：修改配置后重新测试，对比结果
5. **记录结果**：保存压测结果，便于后续对比

---

## 📚 相关文档

- [pprof 使用说明](./PPROF_USAGE.md)
