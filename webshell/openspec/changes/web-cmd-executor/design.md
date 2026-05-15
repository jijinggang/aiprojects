## Context

全新项目，目标是构建一个 Web 端 .cmd 命令执行工具。团队需要统一入口执行服务器上的运维/构建脚本，实时查看输出，多人协作时有互斥保护。当前没有相关代码。

## Goals / Non-Goals

**Goals:**
- 单二进制部署（Go embed 前端资源）
- .cmd 文件级别的互斥执行
- 实时输出推送（所有观看者同步看到 stdout/stderr）
- 刷新页面后恢复完整输出（服务端内存持久化）
- 最近 N 次执行历史记录和回放

**Non-Goals:**
- 命令参数传递（用户不能自定义 .cmd 的参数）
- 终止正在执行的命令
- 用户认证/权限系统
- 跨平台支持（仅 Windows，cmd.exe）
- 输出的 ANSI 颜色码渲染
- 文件变化热更新（启动后不重新扫描目录）
- 持久化到磁盘/数据库（仅内存）

## Decisions

### 1. 纯 WebSocket 通信，无 REST API

**选择**: 所有业务交互通过单一 WebSocket 连接完成，HTTP 仅提供页面（`/`）和 WebSocket 升级（`/ws`）。

**替代方案**: REST API + WebSocket 混合（POST `/api/run` 启动命令，WebSocket 仅推送输出）。

**理由**: 单一通信通道减少前后端协议复杂度，WebSocket 已支持双向消息（run、watch、replay 都能通过 JSON 消息完成）。REST 混合方案需要维护两套请求处理逻辑。

### 2. 输出存储：slice + 行数上限

**选择**: 使用 `[]OutputLine` 切片存储输出行，设置最大行数上限（默认 10000），超限丢弃最早行。

**替代方案**: Ring buffer（环形缓冲区 + seq 定位）。

**理由**: 大多数 .cmd 执行输出不超过几千行，slice 更简单易理解。环形缓冲区的 seq 定位机制增加了复杂度但收益有限。行数上限保证内存可控，丢弃最早行对用户体验影响小（历史回放场景中极少需要查看开头几行）。

### 3. 项目结构：3 个 Go 文件

**选择**:
- `main.go` — 入口、embed、路由、WebSocket handler
- `manager.go` — CmdManager + 数据模型、扫描、互斥、执行、历史
- `output.go` — OutputBuf（输出存储 + 回放）、SubHub（订阅 + 广播）

**替代方案**: 每个概念拆为独立文件（6+ 个 Go 文件）。

**理由**: 项目规模小（约 500-800 行 Go 代码），3 文件足够组织逻辑。过度拆分增加导航成本，不如按职责粗粒度划分。

### 4. 互斥实现：per-.cmd sync.Mutex + running 指针

**选择**: 每个 .cmd 文件有一个 `CmdLock{mu: sync.Mutex, running: *ExecRecord}`，执行时设置 running 指针，结束时置 nil。

**替代方案**: 全局队列或 Redis 分布式锁。

**理由**: 单服务器内存场景，sync.Mutex 最简单高效。running 指针比计数器更直观，能直接获取当前执行的 ExecRecord。分布式锁对单机部署是过度设计。

### 5. 前端：内联 CSS/JS 的单 HTML 文件

**选择**: 所有 CSS 和 JS 内联在 `index.html` 中，通过 `go:embed` 嵌入二进制。

**替代方案**: 独立 CSS/JS 文件 + embed FS 多文件服务。

**理由**: 项目前端代码量小（约 200 行 CSS + 300 行 JS），内联更简洁，只 embed 一个文件，避免多文件 embed 的路径问题。

### 6. WebSocket 库：gorilla/websocket

**选择**: 使用 `github.com/gorilla/websocket`。

**替代方案**: `nhooyr.io/websocket`（现 `github.com/coder/websocket`）。

**理由**: gorilla/websocket 是最成熟的 Go WebSocket 实现，文档和社区示例丰富。虽然已归档但功能稳定，无已知 bug。nhooyr API 更现代但生态较小。

## Risks / Trade-offs

- **[cmd.exe stdout 缓冲延迟]** → cmd.exe 可能对 stdout 进行行缓冲或完全缓冲，导致输出不是即时推送。缓解：Go 端使用 `exec.Command` 的 Pipe 直连，按行读取（`bufio.Scanner`），大多数 .cmd 输出是逐行的。极端情况下（完全缓冲的长运行命令），输出会延迟到进程结束才一次性推送。

- **[慢客户端阻塞]** → 某个 WebSocket 连接接收慢可能导致输出广播延迟。缓解：SubHub 使用 buffered channel（64 条）+ 非阻塞 `select/default`，慢客户端的输出会被丢弃而不阻塞进程输出读取。新连接通过 Replay 从头获取完整输出。

- **[内存占用]** → 大量执行历史或超大输出的命令可能占用较多内存。缓解：每个 ExecRecord 的 OutputBuf 有行数上限（10000 行），历史只保留最近 N 次（默认 5），总内存上限约为 `5 * 10000 * 平均行长度`。

- **[无认证]** → 任何能访问 Web 页面的人都能执行命令。缓解：这是当前 non-goal，后续可添加简单认证。建议通过防火墙/网络隔离限制访问范围。

- **[服务重启丢失状态]** → 所有执行状态和输出仅存内存，重启后全部清空。缓解：这是设计意图，刷新恢复通过内存持久化实现，重启清空是可接受的 trade-off。