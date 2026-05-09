## Why

Windows 环境中，未签名的 EXE/DLL 文件存在安全隐患——可能被篡改、携带恶意代码或不符合企业合规要求。运维和安全团队需要一个轻量工具，批量扫描目录中的未签名 PE 文件并打包，便于审计、上报或进一步分析。目前缺少这样一款开箱即用的 CLI 工具。

## What Changes

- 新增一个 Go 语言命令行工具 `scan_unsigned_files`
- 递归扫描指定目录中的所有 `.exe` / `.dll` 文件
- 调用 Windows `WinVerifyTrust` API 验证数字签名（无签名、签名无效、证书过期/吊销均视为未签名）
- 默认行为：将未签名文件按原始目录结构打包为 ZIP
- 支持 `--list` 模式：仅列出未签名文件路径、大小、未签名原因，不生成 ZIP
- 支持 `-o` 指定 ZIP 输出路径

## Capabilities

### New Capabilities

- `pe-signature-verification`: 递归扫描目录，通过 WinVerifyTrust API 识别未签名/签名无效的 EXE 和 DLL 文件
- `zip-packaging`: 将收集到的未签名文件按原始目录结构压缩为 ZIP

### Modified Capabilities

<!-- 无现有 capability 需要修改 -->

## Impact

- 新增 Go module：`scan_unsigned_files`
- 依赖：`golang.org/x/sys`（Windows syscall 封装）
- 仅支持 Windows 平台编译和运行
