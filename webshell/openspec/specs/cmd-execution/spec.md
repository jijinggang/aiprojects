# cmd-execution Specification

## Purpose
TBD - created by archiving change web-cmd-executor. Update Purpose after archive.
## Requirements
### Requirement: 同一 .cmd 文件同一时间只能被一人执行
系统 SHALL 对每个 .cmd 文件实施互斥控制，同一 .cmd 文件同一时间只允许一个执行实例。不同 .cmd 文件可以并行执行。

#### Scenario: 命令空闲时成功启动
- **WHEN** 用户 A 对空闲的 `build.cmd` 发送 `{type: "run", cmdName: "build.cmd"}`
- **THEN** 系统启动执行，返回 `{type: "status", status: "running"}` 并分配 execId

#### Scenario: 命令已被他人占用时拒绝
- **WHEN** 用户 A 对正在执行的 `build.cmd` 发送 `{type: "run", cmdName: "build.cmd"}`
- **THEN** 系统返回 `{type: "error", code: "CMD_LOCKED", message: "build.cmd 正在执行中"}`

#### Scenario: 不同命令可并行执行
- **WHEN** `build.cmd` 正在执行，用户 B 对空闲的 `deploy.cmd` 发送 `{type: "run", cmdName: "deploy.cmd"}`
- **THEN** `deploy.cmd` 成功启动执行

### Requirement: 执行完成后释放互斥锁
系统 SHALL 在进程结束后（无论成功或失败）立即释放该 .cmd 文件的互斥锁，允许后续执行。

#### Scenario: 命令正常结束后可再次执行
- **WHEN** `build.cmd` 执行完成（exit code 0）
- **THEN** `build.cmd` 的互斥锁释放，其他用户可以启动新的执行

#### Scenario: 命令失败结束后可再次执行
- **WHEN** `build.cmd` 执行失败（exit code 非 0）
- **THEN** `build.cmd` 的互斥锁同样释放

### Requirement: 使用 cmd.exe 执行 .cmd 文件
系统 SHALL 使用 `cmd.exe /c <绝对路径>` 执行 .cmd 文件。

#### Scenario: 执行指定 .cmd 文件
- **WHEN** 系统启动执行 `E:\scripts\build.cmd`
- **THEN** 系统调用 `cmd.exe /c E:\scripts\build.cmd` 作为子进程

### Requirement: 不允许终止正在执行的命令
系统 SHALL NOT 提供终止正在执行命令的能力。命令必须自然结束。

#### Scenario: 用户尝试终止命令
- **WHEN** 用户发送终止请求（若存在）
- **THEN** 系统拒绝该请求

### Requirement: 记录执行元数据
每次执行 SHALL 记录：自增 ID、命令名、开始时间、结束时间（完成后）、状态（running/done/failed）、退出码。

#### Scenario: 执行开始时记录元数据
- **WHEN** `build.cmd` 开始执行
- **THEN** 创建 ExecRecord，包含 ID、cmdName、StartTime、status=running、exitCode=-1

#### Scenario: 执行完成后更新元数据
- **WHEN** `build.cmd` 执行结束（exit code 0）
- **THEN** 更新 ExecRecord 的 EndTime、status=done、exitCode=0

### Requirement: 保留最近 N 次执行历史
系统 SHALL 为每个 .cmd 文件保留最近 N 次执行记录（N 默认 5，可通过 `-max-history` 参数配置）。超出 N 时丢弃最旧的已完成记录。

#### Scenario: 历史记录达到上限
- **WHEN** `build.cmd` 已有 5 条历史记录，第 6 次执行完成
- **THEN** 最旧的已完成记录被移除，当前保留第 2-6 次记录

#### Scenario: 运行中的记录不计入淘汰
- **WHEN** `build.cmd` 有 4 条完成历史 + 1 条正在运行，新的执行启动
- **THEN** 淘汰最旧的已完成记录，运行中的记录不被淘汰

