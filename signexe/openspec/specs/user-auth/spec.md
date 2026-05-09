## ADDED Requirements

### Requirement: 用户认证
系统 SHALL 使用 HTTP Basic Auth 对 Web 界面进行访问控制，用户凭据由配置文件管理。

#### Scenario: 有效凭据登录
- **WHEN** 用户输入正确的用户名和密码
- **THEN** 系统显示 EXE 签名工具主界面

#### Scenario: 无效凭据拒绝
- **WHEN** 用户输入错误的用户名或密码
- **THEN** 系统返回 401 并显示认证失败提示

#### Scenario: 未认证访问拒绝
- **WHEN** 用户未通过认证直接访问页面
- **THEN** 系统弹出浏览器原生登录对话框

### Requirement: 用户列表配置
系统 SHALL 从 config.yaml 的 `auth.users` 列表中加载所有用户凭据。

#### Scenario: 多用户配置
- **WHEN** config.yaml 中配置了多个 username/password 对
- **THEN** 所有用户均可成功登录