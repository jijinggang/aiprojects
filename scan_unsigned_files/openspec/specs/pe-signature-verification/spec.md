### Requirement: 递归扫描目录

系统 SHALL 递归遍历用户指定的目录，收集所有扩展名为 `.exe` 或 `.dll` 的文件（大小写不敏感）。

#### Scenario: 扫描包含嵌套子目录的目录

- **WHEN** 用户执行 `scan_unsigned_files C:\MyApp`，且该目录下包含 `app.exe`、`lib\util.dll`、`data\config.ini`
- **THEN** 系统只收集 `C:\MyApp\app.exe` 和 `C:\MyApp\lib\util.dll`，跳过 `config.ini`

#### Scenario: 扩展名大小写混合

- **WHEN** 目录中包含 `App.EXE` 和 `Helper.DLL`
- **THEN** 系统 SHALL 将其识别为 PE 文件并纳入扫描

### Requirement: 签名验证

系统 SHALL 对每个 PE 文件调用 WinVerifyTrust API（`WINTRUST_ACTION_GENERIC_VERIFY_V2`）验证数字签名。

#### Scenario: 文件有有效签名

- **WHEN** 文件具有有效数字签名（证书链完整、未过期、未吊销）
- **THEN** 系统将其标记为"已签名"，不加入未签名列表

#### Scenario: 文件无数字签名

- **WHEN** 文件没有任何数字签名（TRUST_E_NOSIGNATURE）
- **THEN** 系统将其加入未签名列表，原因为"无数字签名"

#### Scenario: 文件签名无效

- **WHEN** 文件有数字签名但哈希不匹配或签名损坏
- **THEN** 系统将其加入未签名列表，原因为"签名无效"

#### Scenario: 证书已过期

- **WHEN** 文件有数字签名但签名证书已过期
- **THEN** 系统将其加入未签名列表，原因为"证书已过期"

#### Scenario: 证书已吊销

- **WHEN** 文件有数字签名但签名证书已被吊销
- **THEN** 系统将其加入未签名列表，原因为"证书已吊销"

### Requirement: 错误处理

系统 SHALL 在无法访问目录或文件时输出错误信息并继续处理剩余文件。

#### Scenario: 无权限访问子目录

- **WHEN** 扫描过程中遇到无权限访问的子目录
- **THEN** 系统输出警告信息，跳过该子目录，继续扫描其他目录

#### Scenario: 文件读取失败

- **WHEN** 某个 PE 文件无法读取（被锁定、权限不足）
- **THEN** 系统输出错误信息并跳过该文件，继续处理其他文件