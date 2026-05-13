## Context

单文件 Go CLI 工具，从零开始。排除规则通过命令行位置参数传入（无 `.zipignore` 文件），不实现否定模式，路径匹配遵循 gitignore 语义。

## Goals / Non-Goals

**Goals:**
- 接收源目录和排除规则，输出 ZIP 文件
- 排除规则严格遵循 gitignore 语义
- TDD 开发：每个功能先写测试再实现

**Non-Goals:**
- 不支持 `!` 否定模式
- 不读取 `.zipignore` 或 `.gitignore` 文件
- 不支持压缩级别、密码、分卷等高级特性

## Decisions

### 1. `go-gitignore` 库处理模式匹配

**选择**：`github.com/sabhiram/go-gitignore`

**原因**：gitignore 的 `*`、`**`、`?`、字符类、目录锚定等规则复杂，手写匹配引擎成本高且易出错。该库是 Go 生态中 gitignore 的事实标准，API 简单（`CompileIgnoreLines` + `MatchesPath`）。

### 2. 位置参数风格 CLI

**选择**：`ziptool <source_dir> -p [patterns...] [-o output.zip]`

**原因**：排除规则在 gitignore 中本就是一行一个，自然映射为位置参数。

### 3. 目录匹配追加 `/` 后缀

**选择**：WalkDir 遍历中对目录条目匹配路径末尾追加 `/`

**原因**：`go-gitignore` 的 `MatchesPath` 对路径字符串匹配，追加 `/` 模拟 gitignore 中 `build/` 只匹配目录的语义。

### 4. 分离关注点

**选择**：三步管道 — `collectFiles` → `filterPaths` → `writeZip`

**原因**：每个函数职责单一，可独立测试。`writeZip` 只接收最终文件列表，不关心过滤逻辑。

## Risks / Trade-offs

- **[风险] `go-gitignore` 库维护状态**：API 稳定多年 → 缓解：匹配逻辑简单，可随时迁移
- **[风险] 大文件**：使用 `io.Copy` 流式写入，不缓存全部内容
