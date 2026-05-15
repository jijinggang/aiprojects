## ADDED Requirements

### Requirement: 启动时扫描目录发现 .cmd 文件
系统 SHALL 在启动时接收一个目录参数，扫描该目录下所有 `.cmd` 后缀的文件，构建可用命令列表。

#### Scenario: 目录包含多个 .cmd 文件
- **WHEN** 系统以 `-dir E:\scripts` 参数启动，该目录包含 `build.cmd`、`deploy.cmd`、`test.cmd`
- **THEN** 系统发现 3 个 .cmd 文件，命令列表包含这三个文件名

#### Scenario: 目录为空
- **WHEN** 系统以 `-dir E:\empty` 参数启动，该目录不含任何 .cmd 文件
- **THEN** 命令列表为空，系统正常运行（网页显示无可用命令）

#### Scenario: 目录不存在
- **WHEN** 系统以 `-dir E:\nonexistent` 参数启动，该目录不存在
- **THEN** 系统启动失败并报错提示目录不存在

### Requirement: 命令列表包含文件元信息
每个被发现的 .cmd 文件 SHALL 包含文件名、绝对路径、最后修改时间。

#### Scenario: 查看命令列表详情
- **WHEN** 用户通过 WebSocket 连接请求命令列表
- **THEN** 返回每个命令的名称（不含路径）、绝对路径、最后修改时间、当前运行状态、历史记录数量

### Requirement: 命令列表通过 WebSocket 推送
系统 SHALL 通过 WebSocket 在客户端连接时立即推送命令列表，并在客户端请求时重新推送。

#### Scenario: 新客户端连接获取列表
- **WHEN** 新 WebSocket 连接建立
- **THEN** 系统立即推送 `{type: "cmdList"}` 消息，包含所有命令及其状态