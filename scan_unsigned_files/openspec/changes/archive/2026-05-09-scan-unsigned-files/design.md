## Context

一个全新的 Windows CLI 工具，扫描目录中未签名的 PE 文件（EXE/DLL）并打包为 ZIP。工具仅限 Windows 平台运行，利用系统 API `WinVerifyTrust` 判断数字签名有效性。

## Goals / Non-Goals

**Goals:**
- 提供最简单的命令行接口：一个位置参数 + 两个可选 flag
- 准确识别未签名文件（无签名 / 签名无效 / 证书过期或吊销）
- 按原始目录结构输出 ZIP
- 支持仅列出未签名文件清单的模式

**Non-Goals:**
- 不支持跨平台（macOS/Linux）
- 不支持除 EXE/DLL 以外的 PE 文件类型（如 .sys, .ocx）
- 不提供 GUI
- 不支持自定义签名验证策略（如忽略过期证书）
- 不修复或重新签名文件

## Decisions

### D1: 使用 WinVerifyTrust API

**选择**：调用 `wintrust.dll` 的 `WinVerifyTrust` 函数，通过 `golang.org/x/sys/windows` 加载 DLL 和构造参数。

**备选**：纯 Go 解析 PE 头 + PKCS#7 签名 + 证书链验证。

**理由**：WinVerifyTrust 是 Windows 原生签名验证机制，包含完整的证书链验证、吊销检查、时间戳验证。纯 Go 实现工作量巨大且容易遗漏边界情况。工具本身仅处理 Windows PE 文件，跨平台无实际意义。

### D2: PE 文件类型判断基于扩展名而非魔术字节

**选择**：检查文件名后缀 `.exe` / `.dll`（大小写不敏感）。

**理由**：扫描大量文件时检查扩展名比读取文件头更快；实际场景中 PE 文件几乎总是使用标准扩展名。仅在 WinVerifyTrust 返回错误时进一步判断文件格式。

### D3: 四文件结构

**选择**：`main.go` / `scanner.go` / `verify.go` / `archiver.go`，无子包。

**理由**：工具总代码量预计 < 300 行，拆分为子包属于过度设计。四个文件各自职责清晰，无需额外包边界。

### D4: CLI 使用标准库 flag

**选择**：Go 标准库 `flag` 包，不使用 `cobra` 等第三方 CLI 框架。

**理由**：只有 2 个 flag（`-o` 和 `--list`），`flag` 完全够用。零额外依赖是目标之一。

## Risks / Trade-offs

- **[R] WinVerifyTrust 返回结果受系统证书存储影响**：不同机器的证书信任链可能不同 → 可接受，这正是使用 OS 原生验证的目的
- **[R] 扫描大型目录（数万文件）性能**：每个 PE 文件都会触发完整的签名验证，耗时可能较长 → 后续可考虑添加并发扫描，当前版本以简单优先
- **[R] golang.org/x/sys 版本兼容性**：`x/sys` 是实验性包，API 可能变动 → 锁定版本于 go.mod，升级时验证

## Open Questions

<!-- 无 -->
