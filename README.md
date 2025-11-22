# Go Yggdrasil Server

使用 Go 语言 Gin + GORM 框架编写的 Minecraft 登录协议服务端。

## 功能

+ 实现了 Minecraft 登录服务器时的认证部分以及材质部分。支持注册。
+ 兼容 [authlib-injector](https://github.com/yushijinhun/authlib-injector) 。
+ 支持使用在线账号（正版账号）登录，起到透明代理的功能。

## 用途

用于服务器管理员调试和测试时使用小号登录而不必关闭在线验证 (online-mode)。

禁止其他违反 [EULA](https://www.minecraft.net/eula) 的行为。

## 准备

+ 运行 Linux, Windows 或 MacOS 的主机
+ SMTP 服务器和账号用于发送邮箱验证和密码重置邮件（**完全可选**）
  - 无SMTP配置时：用户注册后自动验证，密码重置功能禁用
  - 有SMTP配置时：保持原有邮箱验证流程
+ MySQL 或 PostgreSQL 数据库（如果使用 SQLite 则不需要）

## 用法

下载或编译得到可执行文件并运行，将会自动生成所需的配置文件和数据库文件。

配置文件格式详见 `config_example.ini`，请重命名为 `config.ini` 并放在执行目录下。

启动成功后在启动器（请使用第三方启动器）外置登录选项上填写运行的 URL 的根路径，比如 `http://localhost:8080`。

注册地址在 `/profile/`。

## Docker

使用 docker 快速上手：

```shell
docker run -d --name yggdrasil-go -v $(pwd)/data:/app/data -p 8080:8080 gardel/yggdrasil-go:latest
```

## 实现差异

本实现在完全兼容 [Yggdrasil 服务端技术规范](https://github.com/yushijinhun/authlib-injector/wiki/Yggdrasil-%E6%9C%8D%E5%8A%A1%E7%AB%AF%E6%8A%80%E6%9C%AF%E8%A7%84%E8%8C%83) 的基础上，还提供了额外的扩展功能：

### 核心规范兼容性
**100% 实现** - 所有12个核心API端点均已完全实现，无缺失或重大差异

### 扩展功能（16个额外端点）
1. **用户管理扩展**（5个）: 注册、邮箱验证、密码重置、角色切换
2. **材质管理扩展**（3个）: 材质文件下载、URL设置、重复路径
3. **Minecraft 1.19+兼容**（4个）: 玩家属性、隐私设置、消息签名
4. **Mojang API兼容**（4个）: profile lookup系列端点

### 实现细节差异
1. **离线登录兼容性**（中影响）：未采用与Mojang离线验证兼容的UUID生成方式，可能影响从离线验证系统迁移的用户数据兼容性
2. **材质ID生成算法**（低影响）：使用了与Mojang不同的材质hash计算方法

## 贡献指南

我们欢迎任何形式的贡献！在提交贡献之前，请注意：

### 开发环境准备
1. 安装 Go 1.19+
2. 安装 Node.js 和 Yarn（用于前端开发）
3. Fork 项目并创建功能分支

### 代码规范
- 遵循项目现有的代码风格
- 添加必要的测试用例
- 确保所有测试通过：`go test ./...`
- 运行代码格式化：`go fmt ./...`

### 提交 Pull Request
1. 编写清晰的提交信息
2. 关联相关 Issue（如果有）

## 计划

- [x] 支持密码重置
- [x] 支持不同的数据库如 PostgreSQL 等
- [x] SMTP配置可选化（无SMTP时自动禁用邮箱验证）
- [ ] 添加选项以支持完全离线模式（不检查 Mojang 接口）
- [ ] 令牌持久化防止升级和重启时令牌生效