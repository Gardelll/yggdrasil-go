# Go Yggdrasil Server

使用 Go 语言 Gin + GORM 框架编写的 Minecraft 登录协议服务端。

## 功能

+ 实现了 Minecraft 登录服务器时的认证部分以及材质部分。支持注册。
+ 兼容 [authlib-injector](https://github.com/yushijinhun/authlib-injector) 。
+ 支持使用在线账号（正版账号）登录，起到透明代理的功能。

## 用途

用于服务器管理员调试和测试时使用小号登录而不必关闭在线验证 (online-mode)。

禁止使玩家绕过在线验证登录服务器而不必购买 Minecraft。

禁止其他违反 [EULA](https://account.mojang.com/documents/minecraft_eula) 的行为。

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

## 计划

- [x] 支持密码重置
- [x] 支持不同的数据库如 PostgreSQL 等
- [x] SMTP配置可选化（无SMTP时自动禁用邮箱验证）
- [ ] 添加选项以支持完全离线模式（不检查 Mojang 接口）
- [ ] 令牌持久化防止升级和重启时令牌生效