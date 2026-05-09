## ADDED Requirements

### Requirement: 证书扫描
系统 SHALL 在启动时扫描 `certificates/` 目录下所有 `.pfx` 文件，并提取每个证书的元数据。

#### Scenario: 扫描到多个证书
- **WHEN** certificates 目录包含多个 .pfx 文件
- **THEN** 证书下拉框列出所有证书，显示名称包含主题 CN 和过期日期

#### Scenario: 无证书
- **WHEN** certificates 目录为空或无 .pfx 文件
- **THEN** 证书下拉框为空，界面提示无可选证书

### Requirement: 证书元数据提取
系统 SHALL 使用 certutil 工具提取 PFX 证书的主题、颁发者和过期时间。

#### Scenario: 正常证书解析
- **WHEN** PFX 文件完整且 certutil 可用
- **THEN** CertInfo 包含 subject、issuer、expiry 字段

#### Scenario: 损坏证书处理
- **WHEN** PFX 文件已损坏且 certutil 返回非零退出码
- **THEN** 该证书不出现在下拉列表中，日志记录错误

### Requirement: 证书密码匹配
系统 SHALL 从 config.yaml 的 `certificates.passwords` 中查找指定证书文件名的密码。

#### Scenario: 密码已配置
- **WHEN** 证书文件名在 passwords 映射中存在
- **THEN** 返回对应的密码字符串

#### Scenario: 密码未配置
- **WHEN** 证书文件名在 passwords 映射中不存在
- **THEN** 抛出错误，提示缺少密码配置

### Requirement: 刷新证书列表
系统 SHALL 提供"刷新证书"按钮，触发重新扫描证书目录。

#### Scenario: 手动刷新
- **WHEN** 用户点击刷新按钮
- **THEN** 系统重新扫描证书目录并更新下拉列表