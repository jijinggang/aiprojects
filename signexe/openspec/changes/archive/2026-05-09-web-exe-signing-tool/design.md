## Context

当前项目为空仓库。需要在 Windows Server 上构建一个 Web 版 EXE 签名工具。用户通过浏览器操作，服务器端集中管理证书和 signtool.exe。技术选型：Python + Gradio 5.x，采用 TDD 红/绿测试模式开发。

约束条件：
- 部署环境：Windows Server
- 签名工具：signtool.exe（Windows SDK 自带）
- 证书格式：PFX/P12 文件
- 前端框架：Gradio（类 Gradio 简单方案）
- 认证：简单账号密码，配置文件管理
- 证书安全：直接存服务器文件系统（简化方案）
- ZIP 处理：Python stdlib `zipfile`，无额外依赖

## Goals / Non-Goals

**Goals:**
- 提供 Web UI 完成 EXE/DLL/ZIP 上传→选证书→签名→下载的完整流程
- 支持 ZIP 内批量签名（解压→签名所有 EXE/DLL→重新打包）
- 支持多证书管理与切换
- 签名过程实时流式输出日志
- 仅 2 个运行时依赖，部署简单
- TDD 测试覆盖核心模块

**Non-Goals:**
- 不实现用户注册/自助管理
- 不做审计日志持久化
- 不做分布式/队列/多 worker
- 不支持多用户并发签名
- 不实现证书密码的加密存储（如 Azure Key Vault）
- 不做前端构建（无 npm/webpack）
- 不递归处理嵌套 ZIP 文件（仅处理一层）

## Decisions

### 1. Gradio 5.x 而非 Streamlit 或 FastAPI 自建前端

**理由**: 用户指定"gradio类似方案"。Gradio 内置 auth（`auth=` 参数）、`gr.File()` 上传下载、`gr.Blocks()` 多步骤布局、generator 流式输出，无需写 HTML/JS。

**替代考虑**: Streamlit 更适合数据分析场景，其 auth 需第三方插件。FastAPI + Jinja2 需要更多前端代码。

### 2. Gradio HTTP Basic Auth 而非自定义登录页

**理由**: 最简实现。`auth=[("user", "pass")]` 一行完成。Gradio 5.x 通过 `auth_dependency` 参数控制页面级访问控制。不足：凭证明文 base64 传输，需配合 HTTPS 反向代理。

### 3. PFX 密码存 config.yaml 明文

**理由**: 用户明确要求"不用过多考虑安全问题"。配合 Windows 文件系统 ACL（`icacls` 限制 Administrators 组读取）可满足内网使用场景。

### 4. certutil 提取证书元数据

**理由**: certutil 是 Windows 内置工具，无需额外依赖。不用 `cryptography` 库，保持依赖最小化。

### 5. signtool.exe 自动探测

**理由**: 不同 SDK 版本安装路径不同（`C:\Program Files (x86)\Windows Kits\10\bin\<version>\x64\`）。启动时自动搜索最高版本，找不到时提示明确错误。

### 6. 签名文件加 .signed 后缀而非就地覆盖

**理由**: 保留原始文件以便失败重试。签名完成后由用户决定是否下载。

### 7. 扁平文件结构，无子包

**理由**: 项目小（7 个 py 文件）。每个文件可直接 `import`，无需 `__init__.py` 包管理。

### 8. Python `zipfile` 处理 ZIP 文件

**理由**: `zipfile` 是 Python 标准库，无需额外依赖。流程：解压到 `uploads/` 下的临时目录 → 遍历找出所有 `.exe` `.dll` → 逐个调用 `sign_executable` → 签名后的文件打包为 `.signed.zip` → 清理临时目录。

**替代考虑**: 不递归处理嵌套 ZIP（见 Non-Goals）。使用 `tempfile.TemporaryDirectory` 可能因杀进程未清理，改为显式在 `uploads/` 内创建并在签名后清理。

## Risks / Trade-offs

- **大文件（500MB+）占用内存**: Gradio 将上传文件读入内存后交给 handler。超过 200MB 可能阻塞。缓解：`max_file_size` 限制上传大小，文档说明大文件走 SMB 共享直接放置。
- **signtool 时间戳服务器不可达**: 签名失败，日志显示网络错误。缓解：config.yaml 可更换时间戳 URL。
- **并发用户签名冲突**: signtool 对 PFX 文件加文件锁，并发操作自然失败。缓解：前端 `gr.State` 防止同一会话重复提交，文档说明单实例使用。
- **HTTPS 缺失导致凭证明文传输**: 缓解：README 说明必须配合 Nginx/IIS 反向代理做 TLS 终止。
- **大 ZIP 解压占用磁盘空间**: ZIP 解压后可能大于原始文件。缓解：`max_file_size_mb` 限制上传 ZIP 大小，签完后立即清理临时目录。