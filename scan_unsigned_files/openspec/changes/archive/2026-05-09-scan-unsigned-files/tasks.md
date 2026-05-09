## 1. 签名验证模块（verify.go）

- [x] 1.1 [RED] 编写验证函数接口测试：准备一个已知已签名的系统文件（如 `C:\Windows\System32\notepad.exe`）和一个未签名测试文件，测试 `IsSigned()` 能正确区分已签名/未签名状态
- [x] 1.2 [GREEN] 实现 `verify.go`：加载 `wintrust.dll`，定义 `WINTRUST_DATA` 等结构体，调用 `WinVerifyTrust`，返回签名状态和原因
- [x] 1.3 [REFACTOR] 提取 WinVerifyTrust 错误码到签名原因的映射函数，改善错误信息可读性

## 2. 目录扫描模块（scanner.go）

- [x] 2.1 [RED] 编写扫描函数接口测试：在临时目录创建模拟文件结构（含 .exe/.dll/.txt），验证 `Scan()` 正确过滤扩展名并收集文件信息
- [x] 2.2 [GREEN] 实现 `scanner.go`：定义 `FileInfo` 结构体，实现 `Scan()` 函数用 `filepath.WalkDir` 递归遍历，对每个 PE 文件调用 `IsSigned()`，返回未签名文件列表
- [x] 2.3 [REFACTOR] 改进错误处理：增加对无权限目录的跳过逻辑，增加文件读取失败时的容错处理

## 3. ZIP 打包模块（archiver.go）

- [x] 3.1 [RED] 编写打包函数接口测试：构造 `FileInfo` 列表（含不同子目录的文件），调用 `CreateZip()`，解压后验证目录结构正确
- [x] 3.2 [GREEN] 实现 `archiver.go`：用 `archive/zip` 按 `RelPath` 写入文件，保留原始目录结构，使用 `Deflate` 压缩
- [x] 3.3 [REFACTOR] 处理边界情况：无未签名文件时跳过 ZIP 创建，输出提示信息

## 4. CLI 入口（main.go）

- [x] 4.1 [RED] 编写 CLI 集成测试：模拟命令行参数（默认 ZIP 模式 / `-o` 自定义路径 / `--list` 列表模式），验证输出行为
- [x] 4.2 [GREEN] 实现 `main.go`：解析 flag，编排 scanner → archiver/list 流程，处理 `-o` 默认值逻辑
- [x] 4.3 [REFACTOR] 统一错误消息格式、调整输出排版，确保 `--list` 模式输出列对齐

## 5. 端到端验证

- [x] 5.1 [GREEN] 编译 `go build`，在真实 Windows 目录（如 `C:\Program Files\` 下的某个应用）运行完整扫描+打包流程，验证 ZIP 内容正确
- [x] 5.2 [GREEN] 运行 `--list` 模式验证输出格式，确认文件数量、路径、大小、原因列正确
