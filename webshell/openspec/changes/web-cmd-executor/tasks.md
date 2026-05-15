## 1. 项目初始化

- [x] 1.1 RED: 创建 go.mod + gorilla/websocket 依赖，验证模块可编译
- [x] 1.2 GREEN: 初始化 Go module（`go mod init webshell`），添加 `github.com/gorilla/websocket`，创建空的 main.go 确保 `go build` 通过

## 2. OutputBuf 输出存储与回放 (output.go)

- [x] 2.1 RED: 编写 OutputBuf 测试 — 验证 Write 写入行后 Replay 能返回所有行，测试文件: `output_test.go`
- [x] 2.2 GREEN: 实现 OutputBuf 结构体（lines []OutputLine + maxSize + mu），Write 方法追加行分配 seq，Replay(fromSeq) 返回 seq>=fromSeq 的行
- [x] 2.3 RED: 编写 OutputBuf 测试 — 验证行数超 maxSize 时丢弃最早行
- [x] 2.4 GREEN: 实现 maxSize 淘汰逻辑（Write 时检查 len(lines) > maxSize，移除首行）
- [x] 2.5 REFACTOR: 清理 OutputBuf 代码，确保 Replay 返回切片而非逐行复制

## 3. SubHub 发布订阅 (output.go)

- [x] 3.1 RED: 编写 SubHub 测试 — 验证 Subscribe 后收到 Replay 历史输出，测试文件: `output_test.go`
- [x] 3.2 GREEN: 实现 SubHub（subs map + mu），Subscribe 创建 Subscriber（buffered channel 64），先 Replay 再订阅
- [x] 3.3 RED: 编写 SubHub 测试 — 验证 Broadcast 推送给所有订阅者，慢客户端丢弃不阻塞
- [x] 3.4 GREEN: 实现 Broadcast（非阻塞 select/default 推送），Unsubscribe 关闭 channel
- [x] 3.5 RED: 编写 SubHub 测试 — 验证 CloseAll 关闭所有订阅者 channel
- [x] 3.6 GREEN: 实现 CloseAll（遍历 subs，close 每个 channel，清空 map）

## 4. CmdManager 数据模型与扫描 (manager.go)

- [x] 4.1 RED: 编写 CmdManager.Scan 测试 — 验证扫描目录发现 .cmd 文件，目录不存在时报错，测试文件: `manager_test.go`（需创建临时目录和 .cmd 文件）
- [x] 4.2 GREEN: 实现 CmdManager.Scan（os.ReadDir → 过滤 .cmd → 构建 CmdFile + CmdLock + HistoryQueue maps）

## 5. CmdManager 互斥执行 (manager.go)

- [x] 5.1 RED: 编写 CmdManager.Start 测试 — 验证空闲命令可启动、已被占用返回 CMD_LOCKED 错误、不同命令可并行
- [x] 5.2 GREEN: 实现 CmdManager.Start（CmdLock.mu.Lock → 检查 running → 创建 ExecRecord → 设置 running → 启动 goroutine exec.Command）
- [x] 5.3 RED: 编写 CmdManager.Start 测试 — 验证执行完成后互斥锁释放（进程自然结束后 running=nil）
- [x] 5.4 GREEN: 实现 goroutine 执行逻辑（cmd.Wait → 更新 EndTime/Status/ExitCode → running=nil → 广播 status → Hub.CloseAll）

## 6. CmdManager 历史管理 (manager.go)

- [x] 6.1 RED: 编写历史管理测试 — 验证新执行入历史、超出 maxHistory 淘汰最旧已完成记录、运行中的记录不被淘汰
- [x] 6.2 GREEN: 实现历史 Push/淘汰逻辑（Start 时 Push 到 history，完成时检查超 maxHistory 则淘汰最旧 completed record 并释放 OutputBuf）

## 7. WebSocket handler (main.go)

- [x] 7.1 RED: 编写 WebSocket handler 测试 — 验证客户端连接后收到 cmdList 消息，测试文件: `main_test.go`
- [x] 7.2 GREEN: 实现 WSConn 结构体（readPump + writePump goroutine），连接后推送 cmdList
- [x] 7.3 RED: 编写 WebSocket handler 测试 — 验证 run/watch/replay/history 消息路由正确
- [x] 7.4 GREEN: 实现 readPump 消息路由（json.Unmarshal → switch type → 调用 CmdManager 对应方法）
- [x] 7.5 RED: 编写 WebSocket handler 测试 — 验证进程输出通过 WebSocket 推送给订阅者
- [x] 7.6 GREEN: 实现 output 转发（subscriber.Ch → json.Marshal → sendCh → writePump → conn.WriteMessage）

## 8. HTTP 服务与 embed (main.go)

- [x] 8.1 RED: 编写 HTTP 测试 — 验证 GET / 返回嵌入的 HTML 页面
- [x] 8.2 GREEN: 实现 embed（`//go:embed static/index.html`），serveIndex handler，路由注册（/ → serveIndex，/ws → handleWebSocket），main 函数参数解析（-dir/-port/-max-history）
- [x] 8.3 REFACTOR: 清理 main.go，确保 handler 函数职责清晰

## 9. 前端 UI (static/index.html)

- [x] 9.1 创建 static/index.html 基础骨架 — HTML 结构（左右双栏布局），内联 CSS（深色终端风格）
- [x] 9.2 实现前端 WebSocket 连接与消息处理 — connectWS + handleMessage + 自动重连
- [x] 9.3 实现命令列表渲染 — renderCommandList（显示名称、状态标签、观看者数、历史数）
- [x] 9.4 实现点击交互 — 点击空闲命令发送 run+watch，点击运行中命令发送 watch，收到 CMD_LOCKED 自动切换观看
- [x] 9.5 实现终端输出渲染 — appendTerminalLine（stdout 白色、stderr 红色），自动滚动逻辑
- [x] 9.6 实现历史记录与回放 — 显示历史列表，点击回放发送 replay 消息
- [x] 9.7 实现回放输出批量渲染 — replayToTerminal 处理 replayOutput 消息（批量插入 DOM）

## 10. 集成验证

- [x] 10.1 创建测试 .cmd 文件目录（echo_hello.cmd、slow.cmd 含 timeout、fail.cmd exit 1）
- [x] 10.2 手动验证：启动服务 → 浏览器打开 → 命令列表显示 → 执行命令 → 实时输出
- [ ] 10.3 手动验证：刷新页面 → 输出恢复 → 两个浏览器同时点击同一命令 → 互斥生效
- [ ] 10.4 手动验证：历史记录回放 → stderr 红色显示