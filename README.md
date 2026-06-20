# Go 实战练习

十三个由浅入深的 Go 实战项目，覆盖并发、Web、CLI、数据库、测试、性能分析等核心方向。

每个项目的骨架代码已就位，带有 `// TODO` 标注，**你只需要填空**。

---

## 项目列表

### Part 1 — 并发基础

| # | 项目 | 核心练习点 | 难度 |
|---|------|-----------|------|
| 01 | [file-scanner](./01-file-scanner/) | WaitGroup、buffered channel、goroutine 池 | ★☆☆ |
| 02 | [web-crawler](./02-web-crawler/) | context 取消、sync.Map、fan-out | ★★☆ |
| 03 | [downloader](./03-downloader/) | HTTP Range、分片并发写、进度同步 | ★★☆ |
| 04 | [message-queue](./04-message-queue/) | fan-out 广播、select、优雅关闭 | ★★☆ |
| 05 | [chat-room](./05-chat-room/) | goroutine 生命周期、net.Conn、事件循环 | ★★★ |
| 06 | [task-scheduler](./06-task-scheduler/) | time.Ticker、context 传播、panic 恢复 | ★★★ |
| 07 | [kv-store](./07-kv-store/) | sync.RWMutex、TCP 协议解析、TTL 淘汰 | ★★★ |

### Part 2 — Web、CLI、数据库、测试

| # | 项目 | 核心练习点 | 难度 |
|---|------|-----------|------|
| 08 | [mini-http-framework](./08-mini-http-framework/) | Trie 路由、中间件链、Handler 接口 | ★★☆ |
| 09 | [grpc-service](./09-grpc-service/) | protobuf、gRPC server/client、拦截器 | ★★★ |
| 10 | [cli-tool](./10-cli-tool/) | cobra 子命令、viper 配置、goroutine 池扫描 | ★★☆ |
| 11 | [rest-api-db](./11-rest-api-db/) | database/sql、Repository 模式、迁移 | ★★★ |
| 12 | [testing-practice](./12-testing-practice/) | 表驱动测试、httptest、Benchmark | ★★☆ |
| 13 | [pprof-profiling](./13-pprof-profiling/) | CPU/内存/goroutine 泄漏定位与修复 | ★★★ |

---

## 实现计划

### 第一周 — 基础并发模型

**目标：** 掌握 goroutine 启动/回收、WaitGroup、有缓冲 channel。

#### 01-file-scanner（2–3 天）

- [x] 实现 `worker()`：从 `jobs` 读取目录、列出条目、递归入队子目录
- [x] 实现文件信息收集，将 `Result` 发送到 `results` channel
- [x] 处理 `jobs` channel 满时的背压（hint：带 select 的非阻塞发送）
- [x] 使用 `runtime.NumGoroutine()` 验证没有 goroutine 泄漏
- [x] 加入 `-race` 检测：`go run -race ./01-file-scanner`

**完成标志：** 能扫描 1000+ 文件的目录树，goroutine 数量稳定不增长。

---

### 第二周 — Context 与 fan-out 模式

**目标：** 用 context 控制超时/取消，掌握 sync.Map 并发安全去重。

#### 02-web-crawler（3–4 天）

- [x] 实现 `fetch()`：HTTP GET + 解析 `<a href>` 链接（推荐用 `golang.org/x/net/html`）
- [x] 实现 `worker()`：sync.Map 去重、过滤非同域链接、深度限制
- [x] 处理 `jobs` channel 耗尽时的优雅退出（active worker 计数或额外信号）
- [x] 接入 context 超时取消，所有 HTTP 请求都传入 ctx

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

---

### 第五周 — Web 框架原理

**目标：** 理解 `net/http` Handler 接口，手写路由和中间件，看懂 gin/echo 的设计。

#### 08-mini-http-framework（3–4 天）

- [ ] 实现 `router.go` 的 `node.insert()`：将路径按 `/` 分段插入 trie
- [ ] 实现 `node.search()`：静态节点优先匹配，wildcard 节点捕获参数
- [ ] 实现 `middleware.go` 的 `Logger()`：在 `c.Next()` 前后记录耗时
- [ ] 实现 `middleware.go` 的 `Recovery()`：`defer + recover()` 捕获 panic
- [ ] 实现 `Context.String()`：用 `fmt.Fprintf` 写入响应

**完成标志：** `GET /user/:id` 能正确捕获 id 参数，`/panic` 路由触发后服务器不崩溃。

---

### 第六周 — gRPC

**目标：** 掌握 protobuf 定义接口，理解 gRPC 与 REST 的设计差异，学会拦截器。

#### 09-grpc-service（4–5 天）

**前置步骤（先做这个）：**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
mkdir 09-grpc-service/pb
protoc --go_out=./09-grpc-service/pb --go-grpc_out=./09-grpc-service/pb \
       --proto_path=./09-grpc-service/proto \
       ./09-grpc-service/proto/todo.proto
go get google.golang.org/grpc
# 然后删除 server/main.go 和 client/main.go 顶部的 //go:build ignore
```

- [ ] 实现 `GetTodo`：RLock 读 map，`sql.ErrNoRows` 类比返回 `codes.NotFound`
- [ ] 实现 `ListTodos`：遍历 map，返回 repeated Todo
- [ ] 实现 `UpdateTodo`：Lock 写 map，找不到返回 `codes.NotFound`
- [ ] 实现 `DeleteTodo`：Lock 删 map
- [ ] 完善 `loggingInterceptor`：记录 method、duration、error

**完成标志：** 同时运行 server 和 client，client 能完成 Create → List → Update → Delete 全流程。

---

### 第七周 — CLI 工具

**目标：** 掌握 cobra 子命令体系和 viper 配置加载，写出专业级 CLI。

#### 10-cli-tool（3 天）

- [ ] 在 `ports.go` 的 `runPortScan` 中：收集 `results`，对 open 端口排序后输出
- [ ] 在 `files.go` 的 `runFind` 中：实现 `--size-min` / `--size-max` 过滤
- [ ] 在 `files.go` 的 `runRename` 中：实现非 dry-run 的 `os.Rename`，加冲突检测
- [ ] 添加一个新子命令 `hash`：计算文件的 MD5/SHA256（练习 cobra 子命令添加）

```bash
# 测试命令
go run ./10-cli-tool ports localhost --start 1 --end 1024
go run ./10-cli-tool find . --name "*.go"
go run ./10-cli-tool rename . --pattern "^old_" --replace "new_" --dry-run
```

**完成标志：** 三个子命令均可运行，`--help` 输出格式专业。

---

### 第八周 — REST API + 数据库

**目标：** 掌握 `database/sql` 接口、Repository 模式、事务，把 08 的框架用起来。

#### 11-rest-api-db（4–5 天）

- [ ] 完善 `store/sqlite.go` 的 `ListTodos`、`UpdateTodo`、`DeleteTodo`
- [ ] 完善 `handler/todo.go` 的 `updateTodo`、`deleteTodo`
- [ ] 为 `ListTodos` 添加分页：`?page=1&size=20` 查询参数
- [ ] 用 `database/sql` 事务实现"批量完成所有 todo"接口
- [ ] （进阶）将 `net/http` 的 mux 换成 08 的 mini-http-framework

**测试方式：**
```bash
go run ./11-rest-api-db/
curl -X POST localhost:8081/todos -d '{"title":"buy milk"}'
curl localhost:8081/todos
curl -X PATCH localhost:8081/todos/1 -d '{"done":true}'
curl -X DELETE localhost:8081/todos/1
```

**完成标志：** 五个接口全部实现，重启服务后数据仍然存在（SQLite 持久化）。

---

### 第九周 — 测试

**目标：** 掌握 Go 测试的三种核心模式，养成测试先行的习惯。

#### 12-testing-practice（3 天）

- [ ] 实现 `calculator/calc.go` 的 `Fibonacci()`，让 `TestFibonacci` 通过
- [ ] 完成 `api/handler_test.go` 的 `TestList_Empty`：断言响应体是 `[]` 而非 `null`
- [ ] 完成 `TestMethodNotAllowed`：发送 DELETE /items，期望 405
- [ ] 为 `Fibonacci` 添加递归实现，用 `BenchmarkFibonacci` 对比性能

```bash
go test ./12-testing-practice/...          # 跑所有测试
go test -bench=. ./12-testing-practice/calculator/  # 跑 Benchmark
go test -race ./12-testing-practice/...    # 竞争检测
```

**完成标志：** `go test ./...` 全绿，Benchmark 数据显示迭代实现比递归快 100x+。

---

### 第十周 — 性能分析

**目标：** 用 pprof 找出真实性能问题，建立"测量再优化"的习惯。

#### 13-pprof-profiling（3 天）

代码中预置了三个性能问题，用 pprof 逐一找到并修复：

- [ ] **CPU 热点**：用 CPU profile 定位 `inefficientFib`，改为迭代实现
- [ ] **内存泄漏**：用 heap profile 定位 `leakyCache`，添加最大容量限制
- [ ] **goroutine 泄漏**：用 goroutine profile 找到卡在 `<-ch` 的泄漏，用 context 修复
- [ ] 实现 `/stats` 接口：返回 `runtime.NumGoroutine()` 和 `runtime.MemStats`

```bash
go run ./13-pprof-profiling/
# 另一个终端
curl localhost:6060/leak       # 触发泄漏
go tool pprof http://localhost:6060/debug/pprof/goroutine
# pprof shell 内: top10 / web / list startLeakyWorker
```

**完成标志：** 三个问题全部修复后，反复调用对应接口，goroutine 数量和内存不再持续增长。

---

## 调试工具速查

```bash
# 竞争检测（每个项目都应跑一遍）
go run -race ./<project>/

# 查看 goroutine 数量
import "runtime"; fmt.Println(runtime.NumGoroutine())

# pprof 三件套（在运行中的服务上执行）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10  # CPU
go tool pprof http://localhost:6060/debug/pprof/heap                # 内存
go tool pprof http://localhost:6060/debug/pprof/goroutine           # goroutine

# 运行测试
go test ./...                    # 全部测试
go test -v -run TestXxx ./pkg/   # 指定测试
go test -bench=. -benchmem ./... # Benchmark + 内存分配统计
```

## 推荐顺序

```
Part 1 并发:  01 → 02 → 03 → 04 → 05 → 06 → 07
Part 2 工程:  08 → 10 → 09 → 11 → 12 → 13
```

Part 2 建议先做 08（框架原理）和 10（CLI），再做 09（gRPC 需要额外工具链），最后 11 可以复用 08 的框架——串联起来更有成就感。

每个项目完成后，回顾骨架里的 TODO 注释，思考为什么这样设计而不是另一种写法。
