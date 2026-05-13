## ADDED Requirements

### Requirement: 递归打包目录为 ZIP

系统 SHALL 将指定源目录下的所有文件和子目录递归打包为一个 ZIP 归档文件。

#### Scenario: 多层目录打包
- **WHEN** 源目录包含 `a.txt`、`sub/b.txt`、`sub/deep/c.txt`
- **THEN** 输出的 ZIP 中包含 `a.txt`、`sub/b.txt`、`sub/deep/c.txt` 三条路径

#### Scenario: 空目录
- **WHEN** 源目录为空
- **THEN** 输出一个空 ZIP 文件

#### Scenario: 源目录不存在
- **WHEN** 指定的源目录路径不存在
- **THEN** 以非零退出码退出并输出错误信息

### Requirement: ZIP 条目路径格式

系统 SHALL 确保 ZIP 内所有条目路径使用正斜杠 `/` 作为路径分隔符。

#### Scenario: Windows 路径转换
- **WHEN** 在 Windows 平台打包，文件相对路径为 `sub\b.txt`
- **THEN** ZIP 条目路径记录为 `sub/b.txt`

### Requirement: 空目录条目

系统 SHALL 在 ZIP 中为每个目录显式创建条目（路径以 `/` 结尾）。

#### Scenario: 保留空目录
- **WHEN** 源目录中存在空子目录 `empty/`
- **THEN** ZIP 中包含条目 `empty/`

### Requirement: 输出路径参数

系统 SHALL 通过 `-o` 参数指定输出 ZIP 文件路径。

#### Scenario: 用户指定输出路径
- **WHEN** 执行 `ziptool ./src -o /tmp/out.zip`
- **THEN** 将 ZIP 文件写入 `/tmp/out.zip`

#### Scenario: 未指定输出路径
- **WHEN** 执行 `ziptool ./src` 不带 `-o`
- **THEN** 默认写入当前目录 `output.zip`

### Requirement: Zip Slip 防护

系统 SHALL 拒绝 ZIP 条目路径中包含 `..` 路径遍历片段的条目。

#### Scenario: 检测到路径遍历
- **WHEN** 文件列表中存在相对路径含 `..` 片段的条目
- **THEN** 输出警告并跳过该条目