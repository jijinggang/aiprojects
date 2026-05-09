## ADDED Requirements

### Requirement: 按原始目录结构打包 ZIP

系统 SHALL 将收集到的未签名文件按原始相对目录结构打包为一个 ZIP 文件。

#### Scenario: 保留嵌套目录结构

- **WHEN** 未签名文件为 `C:\scan\app.exe` 和 `C:\scan\plugins\helper.dll`
- **THEN** ZIP 文件内包含 `app.exe`（根目录）和 `plugins/helper.dll`（子目录）

#### Scenario: 默认 ZIP 输出路径

- **WHEN** 用户未指定 `-o` 参数，扫描目录为 `C:\Program Files\MyApp`
- **THEN** 系统在当前工作目录生成 `MyApp_unsigned.zip`

#### Scenario: 自定义 ZIP 输出路径

- **WHEN** 用户指定 `-o D:\result\out.zip`
- **THEN** 系统在 `D:\result\out.zip` 生成 ZIP 文件

#### Scenario: 无未签名文件时

- **WHEN** 扫描目录中所有 PE 文件均已签名
- **THEN** 系统输出提示信息（"未发现未签名文件"），不生成空的 ZIP

### Requirement: 列表输出模式

系统 SHALL 支持 `--list` 参数，仅输出未签名文件列表而不生成 ZIP。

#### Scenario: 列表模式输出

- **WHEN** 用户执行 `scan_unsigned_files --list C:\scan`
- **THEN** 系统中终端显示每个未签名文件的相对路径、文件大小（KB）和未签名原因

#### Scenario: 列表模式无未签名文件

- **WHEN** 用户执行 `--list` 模式且所有文件均已签名
- **THEN** 系统输出"未发现未签名文件"
