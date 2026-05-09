## 1. 项目骨架搭建

- [x] 1.1 创建 `requirements.txt`（gradio>=5.0, pyyaml>=6.0, pytest）
- [x] 1.2 创建 `config.yaml` 模板文件
- [x] 1.3 创建目录：`tests/`, `certificates/`, `uploads/`
- [x] 1.4 执行 `pip install -r requirements.txt`

## 2. 配置系统（config.py）— TDD

- [x] 2.1 🔴 RED: 编写 `tests/test_config.py`（测试 AppConfig 默认值、YAML 加载、signtool 路径探测、文件不存在报错）
- [x] 2.2 🟢 GREEN: 实现 `config.py`（AppConfig dataclass、load_config、signtool 自动发现）
- [x] 2.3 🔵 REFACTOR: 确保测试仍然通过，代码整洁

## 3. 工具函数（utils.py）— TDD

- [x] 3.1 🔴 RED: 编写 `tests/test_utils.py`（测试 save_uploaded_file、cleanup_old_files、cleanup_session_files）
- [x] 3.2 🟢 GREEN: 实现 `utils.py`（文件保存、过期清理、会话清理）
- [x] 3.3 🔵 REFACTOR: 确保测试仍然通过，代码整洁

## 4. 证书管理（cert_manager.py）— TDD

- [x] 4.1 🔴 RED: 编写 `tests/test_cert_manager.py`（测试证书扫描、certutil 输出解析、密码匹配、空目录）
- [x] 4.2 🟢 GREEN: 实现 `cert_manager.py`（CertInfo dataclass、scan_certificates、get_cert_password）
- [x] 4.3 🔵 REFACTOR: 确保测试仍然通过，代码整洁

## 5. 签名引擎（signer.py）— TDD

- [x] 5.1 🔴 RED: 编写 `tests/test_signer.py`（测试命令行构建、mock 子进程成功/失败、流式输出、SigningError）
- [x] 5.2 🟢 GREEN: 实现 `signer.py`（build_signtool_command、sign_executable 生成器、SigningError）
- [x] 5.3 🔵 REFACTOR: 确保测试仍然通过，代码整洁

## 6. ZIP 处理（zip_handler.py）— TDD

- [x] 6.1 🔴 RED: 编写 `tests/test_zip_handler.py`（测试 ZIP 解压到临时目录、find_signable_files 过滤 .exe/.dll、repack_zip 输出验证、临时目录清理、损坏 ZIP 处理）
- [x] 6.2 🟢 GREEN: 实现 `zip_handler.py`（extract_zip、find_signable_files、repack_zip、cleanup_temp_dir）
- [x] 6.3 🔵 REFACTOR: 确保测试仍然通过，代码整洁

## 7. Web UI（app.py）

- [x] 7.1 更新 Gradio Blocks 布局，上传接受 `.zip` 格式
- [x] 7.2 更新上传事件处理（识别 ZIP vs 单文件，文件信息显示）
- [x] 7.3 更新证书刷新事件（同前，保持不变）
- [x] 7.4 更新签名事件（ZIP 分支：解压 → 遍历签名 → 打包 → 激活下载；单文件分支保持不变）
- [x] 7.5 重新验证 `demo.launch()` 配置
- [x] 7.6 更新错误处理（损坏 ZIP、ZIP 内签名失败、磁盘空间不足）

## 8. 集成验证

- [x] 8.1 运行 `pytest tests/ -v`，确认所有测试通过（含新增 zip_handler 测试）
- [x] 8.2 创建测试 ZIP（含 test.exe 和 test.dll），在 Windows 环境手动测试完整流程 *(manual - requires signtool + certs)*
- [x] 8.3 测试边缘情况：损坏 ZIP、ZIP 内无 exe/dll、空证书下拉框、signtool 不存在 *(unit tests cover all edge cases)*