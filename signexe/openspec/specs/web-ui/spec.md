## ADDED Requirements

### Requirement: 文件上传
系统 SHALL 提供文件上传按钮，接受 .exe .dll .msi .zip 等文件格式，保存到 uploads 目录。

#### Scenario: 上传 EXE/DLL 文件
- **WHEN** 用户点击上传按钮并选择 .exe 或 .dll 文件
- **THEN** 文件保存到 uploads 目录，界面显示文件名和大小

#### Scenario: 上传 ZIP 文件
- **WHEN** 用户点击上传按钮并选择 .zip 文件
- **THEN** 文件保存到 uploads 目录，界面显示 "ZIP file" 及文件名和大小

#### Scenario: 上传非允许格式
- **WHEN** 用户尝试上传不允许的文件类型
- **THEN** 上传按钮拒绝该文件

### Requirement: 签名过程实时日志
系统 SHALL 在签名按钮点击后，以实时流式方式在日志区域显示 signtool 输出。

#### Scenario: 流式日志显示
- **WHEN** 用户点击"Sign File"按钮
- **THEN** 日志区域逐行追加 signtool 输出，无需等待签名完成

### Requirement: 签名成功下载
系统 SHALL 在签名成功后激活下载按钮，用户可下载签名后的文件。

#### Scenario: EXE/DLL 签名成功后下载
- **WHEN** signtool 签名成功完成（单文件）
- **THEN** 下载按钮变为可用状态，点击可下载 `.signed.exe` 或 `.signed.dll` 文件

#### Scenario: ZIP 签名成功后下载
- **WHEN** ZIP 内所有文件签名完成并重新打包
- **THEN** 下载按钮变为可用状态，点击可下载 `<原名>.signed.zip` 文件

#### Scenario: 签名失败不激活下载
- **WHEN** signtool 签名失败
- **THEN** 下载按钮保持禁用状态，日志显示错误信息