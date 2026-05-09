## Why

Windows EXE 签名需要安装 Windows SDK、配置证书、记忆 signtool 命令行参数。团队中非开发人员（如运维、发布人员）难以直接操作。提供一个 Web 界面，让用户通过浏览器上传 EXE 即可完成签名并下载，降低使用门槛，集中管理证书。

## What Changes

- 新增 Web UI（Gradio），提供上传→选证书→签名→下载的完整流程
- 新增用户认证（HTTP Basic Auth，账号密码由配置文件管理）
- 新增证书管理模块：扫描服务器端 PFX 证书，展示证书信息（主题、颁发者、过期时间）
- 新增签名引擎：封装 signtool.exe，支持实时流式输出签名日志
- 新增 ZIP 处理模块：解压 ZIP → 签名内部 EXE/DLL → 重新打包为 ZIP 下载
- 新增配置系统：YAML 配置文件管理用户、证书密码、签名参数

## Capabilities

### New Capabilities
- `user-auth`: 简单账号密码认证，由配置文件管理用户列表，Gradio 内置 auth 实现
- `certificate-management`: PFX 证书扫描与元数据提取（通过 certutil），多证书选择
- `exe-signing`: signtool.exe 子进程封装，流式输出签名日志，错误处理。支持 EXE / DLL 签名
- `zip-signing`: ZIP 文件解压，遍历签名内部所有 EXE/DLL，重新打包为 ZIP 供下载
- `web-ui`: Gradio Blocks 构建的多步骤 Web 界面，文件上传/下载，实时日志

### Modified Capabilities
<!-- No existing capabilities to modify -->

## Impact

- 新增文件: `app.py`, `config.py`, `cert_manager.py`, `signer.py`, `zip_handler.py`, `utils.py`, `config.yaml`, `requirements.txt`
- 新增目录: `tests/`, `certificates/`, `uploads/`
- 依赖: `gradio>=5.0`, `pyyaml>=6.0`, 开发依赖 `pytest`。ZIP 处理使用 Python stdlib `zipfile`（无额外依赖）
- 系统依赖: Windows SDK (signtool.exe), certutil (系统自带)