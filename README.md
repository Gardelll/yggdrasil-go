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

## 开服指南

使用本认证服务器开服非常简单，只需以下几个步骤：

### 1. 启动认证服务器
确保认证服务器已在运行，记录下服务器的 URL（例如：`http://localhost:8080`）。

### 2. 下载 authlib-injector
从 [authlib-injector 官网](https://authlib-injector.yushi.moe/) 下载最新版本的 `authlib-injector.jar`。

### 3. 配置 Minecraft 服务端
#### 基本配置
1. 在服务端 `server.properties` 中设置：
   ```
   online-mode=true
   ```
2. 对于 Minecraft 1.19+ 版本，由于本认证服务器兼容正版验证，使用官方启动器的玩家无法新人本认证服务器的签名密钥，需要设置：
   ```
   enforce-secure-profile=false
   ```

#### 启动参数配置
在启动 Minecraft 服务端时添加以下 JVM 参数（参数位于 `-jar` 之前）：

```bash
-javaagent:{path/to/authlib-injector.jar}={your-yggdrasil-server-url} -Dauthlibinjector.disableHttpd
```

请使用 authlibinjector.disableHttpd 参数关闭 authlib-injector 内置的 http 端点，以确保 Minecraft 服务器能直接访问本认证服务器的拓展接口。

例如，如果你将 `authlib-injector.jar` 放在服务端同一目录下，认证服务器运行在 `http://localhost:8080`，则启动命令为：

```bash
java -javaagent:authlib-injector.jar=http://localhost:8080 -Dauthlibinjector.disableHttpd -jar minecraft_server.1.12.2.jar nogui
```

### 4. 玩家登录
玩家需要在支持外置登录的第三方启动器中配置此认证服务器：
1. 选择外置登录模式
2. 填入认证服务器 URL（例如：`http://localhost:8080`）
3. 使用在认证服务器上注册的账号登录

### 5. 使用正版账号登录（不使用authlib-injector）
不使用 authlib-injector 将无法加载外置登录玩家的皮肤，
可以配合此 [修复模组](https://git.gardel.top/magic-server/yggdrasil-skinfix) 使用

### 6. BungeeCord/Velocity 配置（可选）
如果使用 BungeeCord 或 Velocity 代理：
1. 在所有后端服务端和代理服务器上都加载 authlib-injector
2. 只在 BungeeCord/Velocity 上开启 `online-mode=true`
3. 后端服务端设置 `online-mode=false`
4. 在 BungeeCord/Velocity 上开启 `enforce-secure-chat=false`

### 7. 自定义认证服务器上游
在配置文件中可以配置上游认证服务器的端点，也支持配置非 Mojang 官方认证服务器。
但是若用户使用配置的上游认证服务器认证，仅能保证进服，不能保证皮肤和聊天签名可用。
上游认证服务器可在配置中完全关闭，关闭后将完全不验证上游，仅考虑本地账号。

### 注意事项
- 确保认证服务器可通过公共网络访问
- Minecraft 1.19+ 版本必须正确设置 `enforce-secure-profile`
- 本认证服务器兼容正版登录，因此注册时使用的角色名不可与官方服务器重复
- 如果客户端也关闭内置 http 端点，`@mojang` 后缀获取官方皮肤的功能将会失效

## 实现差异

本实现在完全兼容 [Yggdrasil 服务端技术规范](https://github.com/yushijinhun/authlib-injector/wiki/Yggdrasil-%E6%9C%8D%E5%8A%A1%E7%AB%AF%E6%8A%80%E6%9C%AF%E8%A7%84%E8%8C%83) 的基础上，还提供了额外的扩展功能：

### 核心规范兼容性
**100% 实现** - 所有12个核心API端点均已完全实现

### 实现细节差异
1. **离线登录兼容性**（中影响）：未采用与Mojang离线验证兼容的UUID生成方式，可能影响从离线验证系统迁移的用户数据兼容性
2. **材质ID生成算法**（低影响）：使用了与Mojang不同的材质hash计算方法

## 贡献指南

我们欢迎任何形式的贡献！在提交贡献之前，请注意：

### 开发环境准备
1. 安装 Go 1.19+
2. 安装 Node.js 和 Yarn（用于前端开发）
3. Fork 项目并创建功能分支

### 启动后端服务

```shell
go run -tags='sqlite' main.go
```

### 启动前端服务

```shell
cd frontend
yarn dev
```

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
- [x] 添加选项以支持完全离线模式（不检查 Mojang 接口）或多个上游认证服务器
- [ ] 令牌持久化防止升级和重启时令牌生效