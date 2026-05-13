## Why

日常打包项目时需要排除日志文件、编译产物、依赖目录等，手动逐个跳过效率低。需要一个命令行工具，用开发者熟悉的 gitignore 语法指定排除规则。

## What Changes

- 新增 Go CLI 工具 `ziptool`，递归打包目录为 ZIP 文件
- 通过命令行位置参数传入 gitignore 风格的排除规则
- 排除规则遵循 gitignore 语义（不含 `!` 否定模式）

## Capabilities

### New Capabilities

- `zip-archive`: 目录内容递归打包为 ZIP 归档文件
- `gitignore-pattern-match`: 按 gitignore 规则匹配和排除文件路径

### Modified Capabilities

（无）

## Impact

- 新增：`main.go`、`main_test.go`、`go.mod`、`go.sum`
- 依赖：`github.com/sabhiram/go-gitignore`
- 标准库：`archive/zip`、`flag`、`filepath`、`os`、`testing`