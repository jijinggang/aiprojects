# output-streaming Specification

## Purpose
TBD - created by archiving change web-cmd-executor. Update Purpose after archive.
## Requirements
### Requirement: 实时推送 stdout/stderr 输出
系统 SHALL 将正在执行命令的 stdout 和 stderr 输出实时推送给所有订阅该命令的 WebSocket 客户端。

#### Scenario: 执行中的命令产生输出
- **WHEN** `build.cmd` 正在执行，产生 stdout 行 "Compiling...\n" 和 stderr 行 "Warning: deprecated\n"
- **THEN** 所有订阅 `build.cmd` 的 WebSocket 客户端收到两条 `{type: "output"}` 消息，分别标注 stream="stdout" 和 stream="stderr"

### Requirement: 所有人可观看执行中的输出
非执行者 SHALL 也能通过 WebSocket 订阅并观看正在执行的命令的实时输出。

#### Scenario: 非执行者订阅运行中的命令
- **WHEN** 用户 A 启动了 `build.cmd`，用户 B 发送 `{type: "watch", cmdName: "build.cmd"}`
- **THEN** 用户 B 开始接收 `build.cmd` 的实时输出流

### Requirement: 刷新页面后可恢复完整输出
系统 SHALL 在新 WebSocket 连接订阅正在执行的命令时，先回放所有已有输出，然后继续推送实时输出。

#### Scenario: 刷新页面后恢复输出
- **WHEN** `build.cmd` 正在执行，已产生 50 行输出，用户刷新页面后重新发送 `{type: "watch", cmdName: "build.cmd"}`
- **THEN** 用户先收到 50 行历史输出（通过 replayOutput 消息），然后继续接收新输出

### Requirement: 回放已完成命令的输出
系统 SHALL 支持回放已完成命令的历史输出，通过 execId 指定具体执行记录。

#### Scenario: 回放指定历史执行
- **WHEN** 用户发送 `{type: "replay", cmdName: "build.cmd", execId: 42}`
- **THEN** 系统从历史记录中找到 execId=42 的记录，推送完整输出

#### Scenario: 回放已淘汰的历史记录
- **WHEN** 用户发送 `{type: "replay", cmdName: "build.cmd", execId: 10}`，但该记录已被淘汰
- **THEN** 系统返回 `{type: "error", code: "REPLAY_NOT_FOUND"}`

### Requirement: 输出行数上限控制内存
每个 ExecRecord 的输出缓冲区 SHALL 设置最大行数上限（默认 10000）。超限时丢弃最早行。

#### Scenario: 输出超过行数上限
- **WHEN** 命令产生第 10001 行输出
- **THEN** 最早的一行被丢弃，缓冲区保留第 2-10001 行

### Requirement: 进程结束后通知订阅者
系统 SHALL 在进程结束时向所有订阅者发送状态变更消息（done 或 failed），并关闭订阅通道。

#### Scenario: 命令正常结束后推送状态
- **WHEN** `build.cmd` 正常结束（exit code 0）
- **THEN** 所有订阅者收到 `{type: "status", status: "done", exitCode: 0}`，订阅通道关闭

### Requirement: 慢客户端不阻塞输出读取
系统 SHALL 使用非阻塞方式推送输出给慢客户端。如果客户端接收缓冲区满，丢弃该客户端的输出推送，但不阻塞进程输出读取和其他客户端。

#### Scenario: 慢客户端缓冲区满
- **WHEN** WebSocket 客户端 A 接收缓慢，其缓冲 channel 满（64 条未消费）
- **THEN** 客户端 A 的后续输出推送被丢弃，但进程输出读取和其他客户端不受影响

