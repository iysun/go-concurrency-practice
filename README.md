# Go 并发实战练习

七个由浅入深的 Go 实战项目，专项训练 goroutine、channel 及并发原语的实际运用。

每个项目的骨架代码已就位，带有 `// TODO` 标注，**你只需要填空**。

---

## 项目列表

| # | 项目 | 核心练习点 | 难度 |
|---|------|-----------|------|
| 01 | [file-scanner](./01-file-scanner/) | WaitGroup、buffered channel、goroutine 池 | ★☆☆ |
| 02 | [web-crawler](./02-web-crawler/) | context 取消、sync.Map、fan-out | ★★☆ |
| 03 | [downloader](./03-downloader/) | HTTP Range、分片并发写、进度同步 | ★★☆ |
| 04 | [message-queue](./04-message-queue/) | fan-out 广播、select、优雅关闭 | ★★☆ |
| 05 | [chat-room](./05-chat-room/) | goroutine 生命周期、net.Conn、事件循环 | ★★★ |
| 06 | [task-scheduler](./06-task-scheduler/) | time.Ticker、context 传播、panic 恢复 | ★★★ |
| 07 | [kv-store](./07-kv-store/) | sync.RWMutex、TCP 协议解析、TTL 淘汰 | ★★★ |

---

## 实现计划

### 第一周 — 基础并发模型

**目标：** 掌握 goroutine 启动/回收、WaitGroup、有缓冲 channel。

#### 01-file-scanner（2–3 天）

- [ ] 实现 `worker()`：从 `jobs` 读取目录、列出条目、递归入队子目录
- [ ] 实现文件信息收集，将 `Result` 发送到 `results` channel
- [ ] 处理 `jobs` channel 满时的背压（hint：带 select 的非阻塞发送）
- [ ] 使用 `runtime.NumGoroutine()` 验证没有 goroutine 泄漏
- [ ] 加入 `-race` 检测：`go run -race ./01-file-scanner`

**完成标志：** 能扫描 1000+ 文件的目录树，goroutine 数量稳定不增长。

---

### 第二周 — Context 与 fan-out 模式

**目标：** 用 context 控制超时/取消，掌握 sync.Map 并发安全去重。

#### 02-web-crawler（3–4 天）

- [ ] 实现 `fetch()`：HTTP GET + 解析 `<a href>` 链接（推荐用 `golang.org/x/net/html`）
- [ ] 实现 `worker()`：sync.Map 去重、过滤非同域链接、深度限制
- [ ] 处理 `jobs` channel 耗尽时的优雅退出（active worker 计数或额外信号）
- [ ] 接入 context 超时取消，所有 HTTP 请求都传入 ctx

**完成标志：** 爬取 `https://golang.org` 到深度 2，能在 30 秒超时内正常退出。

#### 03-downloader（2–3 天）

- [ ] 实现 `getContentLength()`：HEAD 请求 + `Content-Length` 解析
- [ ] 实现 `downloadChunk()`：`Range: bytes=N-M` 请求 + WriteAt 写入正确偏移
- [ ] 实现 `progressWriter`：包装 Writer，每次 Write 更新进度条
- [ ] 验证下载文件 MD5 与原始文件一致

**完成标志：** 下载一个 100 MB 文件，速度明显快于单线程，进度实时显示。

---

### 第三周 — 事件驱动与 TCP 服务

**目标：** 理解单一事件循环消除锁的模式，掌握 goroutine-per-connection。

#### 04-message-queue（2 天）

- [ ] 实现 `Subscribe()`：创建 Subscriber，注册到 map
- [ ] 实现 `Unsubscribe()`：从 map 删除，关闭 channel
- [ ] 实现 `Close()`：关闭所有 subscriber channel，让消费者 range 循环退出
- [ ] 用 `go test -race` 验证无数据竞争

**完成标志：** 5 个 subscriber 并发消费，任意 subscriber 退出不影响其他人。

#### 05-chat-room（3–4 天）

- [ ] 实现 `Room.Run()` 的 leave 分支：删除 client、关闭 send channel
- [ ] 实现 `writePump()`：range send channel，写入 conn，出错则返回
- [ ] 加入 `/quit` 命令让用户主动断开
- [ ] 加入 `/list` 命令显示当前在线用户数
- [ ] 用 `telnet localhost 9000` 多窗口测试

**完成标志：** 3 个终端同时聊天，任一断开不影响其他连接。

---

### 第四周 — 高级并发与生产级特性

**目标：** panic 安全、优雅关闭、持久化。

#### 06-task-scheduler（3 天）

- [ ] 为每次任务执行创建带超时的子 context
- [ ] 用 `recover()` 捕获任务 panic，记录日志但不崩溃
- [ ] 实现 `Unregister()`：停止指定任务的 ticker
- [ ] 添加任务统计：LastRun、RunCount、ErrorCount

**完成标志：** 注册一个会 panic 的任务，调度器不崩溃并继续运行其他任务。

#### 07-kv-store（4–5 天）

- [ ] 完善 `DEL` 和 `KEYS` 命令处理
- [ ] 在 `SET` 命令中解析可选的 TTL 参数
- [ ] 添加 `EXPIRE key seconds` 命令
- [ ] 添加 `EXISTS key` 命令
- [ ] （进阶）实现 AOF 追加写持久化：每次写操作追加到 `aof.log`，启动时重放

**完成标志：** 用 `telnet` 或 `nc` 能完整执行 SET/GET/DEL/EXPIRE，TTL 到期自动淘汰。

---

## 调试技巧

```bash
# 检测数据竞争（每个项目都应跑一遍）
go run -race ./<project>/

# 查看当前 goroutine 数量（在代码里加）
fmt.Println(runtime.NumGoroutine())

# pprof goroutine 分析（适合 05~07 这类常驻服务）
import _ "net/http/pprof"
go http.ListenAndServe(":6060", nil)
# 然后访问 http://localhost:6060/debug/pprof/goroutine?debug=1
```

## 推荐顺序

```
01 → 02 → 03 → 04 → 05 → 06 → 07
```

每个项目完成后，回顾一次骨架代码里的 TODO 注释，确保理解为什么这样设计而不是另一种写法。
