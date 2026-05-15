## Why

团队共享服务器上的运维/构建 .cmd 脚本，缺乏统一的 Web 执行入口。当前只能通过远程桌面或 SSH 执行，无法实时查看输出，多人操作时容易冲突。需要一个 Web 工具让团队成员通过浏览器选择并执行 .cmd 文件，实时查看输出，刷新不丢失进度。

## What Changes

- 新增 Go 后端服务：启动时扫描指定目录的 .cmd 文件，提供 WebSocket 通信
- 新增 Web 前端页面：命令列表 + 实时输出终端，单页面应用（HTML 内联 CSS/JS）
- 支持 .cmd 文件级别互斥：同一命令同一时间只能被一人执行，不同命令可并行
- 支持实时输出流：所有人均可观看正在执行的命令的实时 stdout/stderr 输出
- 支持刷新恢复：服务端内存持久化输出，刷新页面后可回放完整输出
- 支持历史记录：保留最近 N 次执行记录（默认 5），可回放历史输出
- 单二进制部署：Go embed 前端资源，编译后一个 exe 即可运行

## Capabilities

### New Capabilities
- `cmd-discovery`: 扫描目录发现 .cmd 文件，维护可用命令列表
- `cmd-execution`: 执行 .cmd 文件，互斥控制，进程管理（启动、等待、输出收集）
- `output-streaming`: 实时输出推送（stdout/stderr），历史回放，多订阅者广播
- `web-ui`: 前端单页面界面，命令列表、执行交互、终端输出显示、历史浏览

### Modified Capabilities

（无，全新项目）

## Impact

- 新项目，无现有代码受影响
- 外部依赖：`github.com/gorilla/websocket`（唯一第三方依赖）
- 运行环境：Windows 服务器，需 `cmd.exe` 执行 .cmd 文件
- 部署方式：单二进制 `webshell.exe`，命令行参数 `-dir`、`-port`、`-max-history`