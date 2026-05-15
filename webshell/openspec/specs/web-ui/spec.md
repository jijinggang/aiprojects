# web-ui Specification

## Purpose
TBD - created by archiving change web-cmd-executor. Update Purpose after archive.
## Requirements
### Requirement: 前端单页面包含命令列表和输出终端
系统 SHALL 提供一个单 HTML 页面（内联 CSS + JS），包含左侧命令列表面板和右侧输出终端面板。

#### Scenario: 页面加载显示命令列表
- **WHEN** 用户浏览器打开 `http://localhost:8080/`
- **THEN** 页面显示命令列表面板（左侧）和空输出终端面板（右侧）

### Requirement: 命令列表显示运行状态和历史数量
每个命令 SHALL 显示其当前状态（就绪/运行中）、观看者数量、已完成历史数量。

#### Scenario: 命令正在执行
- **WHEN** `build.cmd` 正在执行
- **THEN** 命令列表中 `build.cmd` 显示"运行中"状态标签和当前观看者数量

#### Scenario: 命令空闲
- **WHEN** `deploy.cmd` 当前无人执行
- **THEN** 命令列表中 `deploy.cmd` 显示"就绪"状态和已完成历史数量

### Requirement: 点击空闲命令可启动执行
用户 SHALL 能通过点击空闲命令来启动执行。

#### Scenario: 点击空闲命令
- **WHEN** 用户点击空闲的 `deploy.cmd`
- **THEN** 前端发送 `{type: "run", cmdName: "deploy.cmd"}` WebSocket 消息

#### Scenario: 点击运行中的命令
- **WHEN** 用户点击正在执行的 `build.cmd`
- **THEN** 前端自动发送 `{type: "watch", cmdName: "build.cmd"}` 订阅输出

### Requirement: 执行被拒绝时提示用户
前端 SHALL 在收到 CMD_LOCKED 错误时提示用户该命令正在被他人执行，并自动切换为观看模式。

#### Scenario: 执行被互斥拒绝
- **WHEN** 用户尝试运行已被他人执行的 `build.cmd`，收到 CMD_LOCKED 错误
- **THEN** 前端显示提示"该命令正在执行中"，自动订阅该命令输出

### Requirement: 输出终端实时显示 stdout/stderr
输出终端 SHALL 实时显示命令的 stdout（白色文字）和 stderr（红色文字）输出。

#### Scenario: 混合 stdout/stderr 输出
- **WHEN** 命令同时产生 stdout 和 stderr 输出
- **THEN** stdout 行显示为白色文字，stderr 行显示为红色文字

### Requirement: 输出终端自动滚动
输出终端 SHALL 自动滚动到底部显示最新输出，除非用户手动向上滚动查看历史。

#### Scenario: 新输出到达时自动滚动
- **WHEN** 新输出行到达且用户未手动向上滚动
- **THEN** 终端自动滚动到底部

#### Scenario: 用户向上滚动时停止自动滚动
- **WHEN** 用户向上滚动查看历史输出
- **THEN** 终端不自动滚动到底部，直到用户手动滚回底部

### Requirement: 可查看和回放历史执行
前端 SHALL 在选中命令时显示其历史执行列表，点击历史条目可回放完整输出。

#### Scenario: 选中命令查看历史
- **WHEN** 用户选中 `build.cmd`
- **THEN** 前端发送 `{type: "history", cmdName: "build.cmd"}` 并显示历史列表

#### Scenario: 点击历史条目回放输出
- **WHEN** 用户点击 `build.cmd` 的历史记录 #42
- **THEN** 前端发送 `{type: "replay", cmdName: "build.cmd", execId: 42}` 并在终端显示完整输出

### Requirement: WebSocket 断线自动重连
前端 SHALL 在 WebSocket 连接断开时自动尝试重连（2 秒间隔），重连后恢复当前观看状态。

#### Scenario: 连接断开后自动重连
- **WHEN** WebSocket 连接意外断开
- **THEN** 2 秒后前端自动尝试重连，重连成功后重新订阅当前正在观看的命令

### Requirement: 前端通过 go:embed 嵌入二进制
所有前端资源（HTML、CSS、JS）SHALL 内联在单个 `index.html` 文件中，通过 `go:embed` 嵌入 Go 二进制。

#### Scenario: 编译后单二进制包含前端
- **WHEN** 执行 `go build -o webshell.exe`
- **THEN** 生成的 `webshell.exe` 包含前端 HTML/CSS/JS，无需额外文件

