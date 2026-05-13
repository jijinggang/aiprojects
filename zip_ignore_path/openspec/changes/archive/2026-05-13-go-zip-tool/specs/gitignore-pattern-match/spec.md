## ADDED Requirements

### Requirement: CLI 位置参数传入排除规则

系统 SHALL 通过位置参数接受 gitignore 风格的排除规则，每个参数对应 `.gitignore` 中的一行。

#### Scenario: 传入多条规则
- **WHEN** 执行 `ziptool ./src -p "*.log" "build/" "temp"`
- **THEN** 三条规则分别排除 .log 文件、build 目录、名为 temp 的条目

#### Scenario: 无排除规则
- **WHEN** 执行 `ziptool ./src` 不带排除规则
- **THEN** 打包所有文件

### Requirement: `*` 和 `?` 通配符

系统 SHALL 支持 `*`（匹配除 `/` 外任意字符序列）和 `?`（匹配除 `/` 外单个字符）。

#### Scenario: 扩展名匹配
- **WHEN** 规则为 `*.log`
- **THEN** 排除所有目录下的 `.log` 文件

#### Scenario: 单字符匹配
- **WHEN** 规则为 `temp?.txt`
- **THEN** 排除 `temp1.txt`、`tempa.txt`，不排除 `temp12.txt`

### Requirement: `**` 双星号通配符

系统 SHALL 支持 `**` 匹配零个或多个目录层级。

#### Scenario: 匹配任意深度
- **WHEN** 规则为 `**/backup`
- **THEN** 排除根下 `backup`、`a/backup`、`a/b/backup`

#### Scenario: 中间任意层级
- **WHEN** 规则为 `a/**/b`
- **THEN** 排除 `a/b`、`a/x/b`、`a/x/y/b`

### Requirement: 目录专用匹配

系统 SHALL 在以 `/` 结尾的规则中仅匹配目录。

#### Scenario: 排除目录不排文件
- **WHEN** 规则为 `build/`，同时存在文件 `build` 和目录 `build/`
- **THEN** 只排除目录，文件 `build` 仍被包含

### Requirement: 根目录锚定

系统 SHALL 支持以 `/` 开头的规则仅匹配根目录。

#### Scenario: 仅排除根下文件
- **WHEN** 规则为 `/config.yml`
- **THEN** 排除根下 `config.yml`，不排除 `sub/config.yml`

### Requirement: 含路径分隔符的规则

系统 SHALL 将包含 `/`（非仅开头或末尾）的规则相对于根目录匹配。

#### Scenario: 精确路径匹配
- **WHEN** 规则为 `sub/logs`
- **THEN** 排除 `sub/logs`，不排除 `other/sub/logs`

### Requirement: 字符类

系统 SHALL 支持 `[...]` 字符类匹配单个字符。

#### Scenario: 字符范围
- **WHEN** 规则为 `*.[oa]`
- **THEN** 排除 `.o` 和 `.a` 文件，不排除 `.txt` 文件

### Requirement: 非法规则报错

系统 SHALL 在排除规则语法无法解析时输出错误并退出。

#### Scenario: 格式错误
- **WHEN** 某条规则语法不合法
- **THEN** 输出错误信息并以非零退出码退出
