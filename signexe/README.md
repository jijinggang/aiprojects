# EXE Signing Tool

基于 Gradio 的 Web 版 Windows EXE 签名工具。通过浏览器上传可执行文件，选择 PFX 证书，调用 signtool.exe 完成数字签名并下载。

## 环境要求

- **操作系统**: Windows Server / Windows 10+
- **Python**: 3.10+
- **系统工具**:
  - [Windows SDK](https://developer.microsoft.com/windows/downloads/windows-sdk/) — 提供 signtool.exe
  - certutil — 系统自带，用于读取证书信息

## 安装

```bash
pip install -r requirements.txt
```

## 配置

编辑 `config.yaml`：

```yaml
auth:
  users:
    - username: admin
      password: your_password

certificates:
  passwords:
    my_cert.pfx: "cert_password"
    another.pfx: "another_password"

signing:
  timestamper_url: "http://timestamp.digicert.com"
  additional_flags: ""          # 可选，如 "/as"
  max_file_size_mb: 500

server:
  host: "127.0.0.1"
  port: 7860

uploads:
  retention_hours: 24
```

- **auth.users**: 用于 Web 登录的账号密码列表
- **certificates.passwords**: `文件名: 密码` 映射，文件名需与 `certificates/` 目录中的 PFX 文件一致
- **signing.timestamper_url**: 时间戳服务器地址（可更换为其他 RFC 3161 服务）
- **signing.additional_flags**: 传递给 signtool 的额外参数（如 `/as` 追加签名）
- **server**: Gradio 监听地址和端口
- **uploads.retention_hours**: 上传文件的保留时间（超时自动清理）

## 使用

### 1. 准备证书

将 `.pfx` 证书文件放入 `certificates/` 目录：

```
certificates/
├── my_company_2025.pfx
└── test_signing.pfx
```

### 2. 启动服务

```bash
python app.py
```

### 3. 访问 Web 界面

打开浏览器访问 `http://127.0.0.1:7860`，输入配置的用户名和密码登录。

### 4. 签名流程

1. 点击 **Upload File** 上传 .exe / .dll / .msi 文件
2. 点击 **Refresh Certificates** 扫描证书
3. 从下拉框选择签名证书
4. 点击 **Sign File** 执行签名，日志区域实时显示 signtool 输出
5. 签名成功后，点击 **Download Signed File** 下载已签名文件

## 生产部署建议

- **HTTPS**: 配合 Nginx / IIS 反向代理做 TLS 终止，避免 Basic Auth 凭证明文传输
- **权限控制**: 使用 `icacls` 限制 `config.yaml` 和 `certificates/` 目录仅 Administrators 可读
- **文件清理**: 定时清理 `uploads/` 目录或配置更短的 `retention_hours`
- **并发限制**: 本工具为单实例设计，不支持多用户并发签名

## 开发

```bash
# 运行所有测试
pytest tests/ -v

# 运行单个模块测试
pytest tests/test_config.py -v
pytest tests/test_signer.py -v
```

项目采用 **TDD（红/绿/重构）** 模式开发，核心模块测试覆盖：
- `test_config.py` — 配置加载、signtool 自动发现
- `test_utils.py` — 文件保存、过期清理
- `test_cert_manager.py` — 证书扫描、certutil 输出解析
- `test_signer.py` — signtool 命令构建、流式签名执行
