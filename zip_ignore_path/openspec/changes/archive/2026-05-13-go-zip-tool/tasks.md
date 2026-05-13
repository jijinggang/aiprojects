## 1. 项目初始化

- [x] 1.1 `go mod init ziptool`，`go get github.com/sabhiram/go-gitignore`

## 2. CLI 参数解析

- [x] 2.1 **RED** — 写 `TestParseFlags`：验证 `-o` 默认值、`sourceDir` 必填校验、patterns 提取并验证 `-p` 默认值(.svn/)、
- [x] 2.2 **GREEN** — 实现 `parseFlags()` 使测试通过

## 3. 目录遍历

- [x] 3.1 **RED** — 写 `TestCollectFiles`：创建临时目录结构 `a.txt, b.log, sub/c.txt, sub/d.log`，遍历收集相对路径
- [x] 3.2 **GREEN** — 实现 `collectFiles(sourceDir string) ([]string, error)` 使测试通过

## 4. gitignore 模式匹配过滤

- [x] 4.1 **RED** — 写 `TestFilterPaths`：用 `["*.log", "sub/"]` 过滤文件列表，验证 .log 文件和 sub 目录被排除
- [x] 4.2 **GREEN** — 实现 `filterPaths(paths []string, patterns []string) ([]string, error)` 使测试通过
- [x] 4.3 **REFACTOR** — 补充测试覆盖 `*`、`?`、`**`、`/` 锚定、末尾 `/` 目录匹配、`[...]` 字符类

## 5. ZIP 文件生成

- [x] 5.1 **RED** — 写 `TestWriteZip`：用固定文件列表生成 zip，`archive/zip` 读取验证条目名称和内容
- [x] 5.2 **GREEN** — 实现 `writeZip(sourceDir, outputPath string, paths []string) error` 使测试通过
- [x] 5.3 **REFACTOR** — 确认空目录条目（`/` 结尾）、Windows 路径转 `/` 分隔符

## 6. Zip Slip 防护

- [x] 6.1 **RED** — 写 `TestZipSlipPrevention`：文件列表含 `../escape.txt`，验证被拒绝
- [x] 6.2 **GREEN** — 在 `writeZip` 中添加 `..` 检测逻辑使测试通过
