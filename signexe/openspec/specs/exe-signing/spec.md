## ADDED Requirements

本 spec 涵盖 EXE、DLL 和 ZIP 文件的签名需求。

### Requirement: signtool 命令行构建
系统 SHALL 根据输入文件和证书信息构建完整的 signtool 命令行参数。适用于 EXE 和 DLL 文件。

#### Scenario: 基本签名命令
- **WHEN** 提供 EXE 路径、证书路径、密码和配置
- **THEN** 命令列表包含 `/f <证书路径>` `/p <密码>` `/fd SHA256` `/tr <时间戳URL>` `/td SHA256` `/v`

#### Scenario: 附加参数
- **WHEN** config.yaml 中 `additional_flags` 包含 `/as`
- **THEN** `/as` 参数出现在 sign 子命令之后

### Requirement: 签名执行
系统 SHALL 通过子进程调用 signtool.exe 执行签名，并以生成器形式流式返回日志行。

#### Scenario: 签名成功
- **WHEN** signtool 正常完成且返回码为 0
- **THEN** 逐行 yield 日志输出，最后 yield 成功消息

#### Scenario: 签名失败
- **WHEN** signtool 返回非零退出码
- **THEN** 抛出 SigningError 异常，包含退出码和完整日志

### Requirement: signtool 自动发现
系统 SHALL 在未显式配置路径时，自动在已知 Windows SDK 路径和 PATH 中搜索 signtool.exe。

#### Scenario: SDK 路径找到
- **WHEN** `C:\Program Files (x86)\Windows Kits\10\bin\` 下存在 signtool.exe
- **THEN** 选择最高版本的 signtool.exe 路径

#### Scenario: 未找到
- **WHEN** 所有探测路径和 PATH 中都找不到 signtool.exe
- **THEN** 应用启动时报错退出，提示已探测的路径列表

### Requirement: ZIP 文件批量签名
系统 SHALL 解压 ZIP 文件，对其内部所有 .exe 和 .dll 文件执行签名，然后将签名后的文件重新打包为 ZIP。

#### Scenario: ZIP 解压并签名
- **WHEN** 上传 .zip 文件包含 setup.exe 和 helper.dll
- **THEN** 文件被解压到临时目录，setup.exe 和 helper.dll 分别被签名，其他非签名目标文件原样保留

#### Scenario: 签名后重新打包
- **WHEN** ZIP 内所有 EXE/DLL 签名完成
- **THEN** 临时目录内所有文件（含未签名的）打包为 `<原名>.signed.zip`，临时目录被清理

#### Scenario: ZIP 内无签名目标
- **WHEN** ZIP 内仅含 .txt .xml 等非 exe/dll 文件
- **THEN** 这些文件原样保留在输出 ZIP 中（无签名操作但仍生成 .signed.zip）

#### Scenario: 签名失败时清理
- **WHEN** ZIP 内某个文件签名失败（signtool 返回非零退出码）
- **THEN** 停止后续签名，临时目录被清理，错误信息记入日志，下载按钮保持禁用

#### Scenario: 损坏的 ZIP 文件
- **WHEN** 上传的文件不是有效的 ZIP 格式
- **THEN** 日志显示 "Invalid ZIP file" 错误，下载按钮保持禁用