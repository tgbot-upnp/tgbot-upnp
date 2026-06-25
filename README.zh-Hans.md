# tgbot-upnp

tgbot-upnp 是一个可以将 Telegram 视频通过 UPnP 协议串流投屏至其他设备的机器人，无需下载完整文件。

[English](https://github.com/tgbot-upnp/tgbot-upnp/blob/main/README.md) | [简体中文](https://github.com/tgbot-upnp/tgbot-upnp/blob/main/README.zh-Hans.md)

## 特性

- 将 Telegram 视频串流投屏到支持 UPnP (AVTransport) 的设备
- 无需下载整个视频，支持进度条
- 自动扫描局域网中所有 UPnP 设备
- 多语言支持（英语、简体中文）
- 低内存和 CPU 占用
- 内置凭证预设 — 无需注册 Telegram API
- 浏览器配置页面，首次运行自动打开
- 系统托盘菜单（Windows / macOS）
- 扫码登录，访问私有频道视频
- 支持消息链接投屏（`t.me/频道名/消息ID`），含私有频道
- 自动管理员：扫码用户自动获得管理权限
- 支持 Docker 部署，数据持久化，多架构镜像

## 快速开始

### 推荐方式：内置凭证

1. 在 [@BotFather](https://t.me/BotFather) 创建机器人 → 获取 `API Token`
2. 在 [@userinfobot](https://t.me/userinfobot) 获取你的用户 ID
3. 从 [releases](https://github.com/tgbot-upnp/tgbot-upnp/releases) 下载运行
4. 自动打开浏览器配置页面 → 选择「Telegram Desktop（官方）」使用内置凭证，填入 Bot Token 和管理员 ID 即可

### Windows 手动配置

编辑 `config.yml`：

```yaml
app_id: 2040                          # 或去 my.telegram.org/apps 申请
api_hash: b18441a1ff607e10a989891a5462e627
bot_token: 123456:ABC-DEF1234...
admin_id: 123456789                   # 多个用英文逗号分隔；开启自动管理员时可不填
auto_admin: true                      # 扫码登录用户自动设为管理员
http_port: 8080
base_url: ""                          # 可选，反向代理地址如 http://your-proxy.com:8080
user_session: ""                      # 扫码登录后自动填充
```

### macOS 运行

从 [releases](https://github.com/tgbot-upnp/tgbot-upnp/releases) 下载 `.app` 包，移至 `/Applications`。

首次打开：右键 `tgbot-upnp.app` → 打开 → 确认。只需一次（Gatekeeper 对未签名应用的提示）。

配置和会话文件存储在 `~/Library/Application Support/tgbot-upnp/`。

### Docker 运行

```shell
docker run -d --name tgbot-upnp \
    -e TGBOT_UPNP_APP_ID="2040" \
    -e TGBOT_UPNP_API_HASH="b18441a1ff607e10a989891a5462e627" \
    -e TGBOT_UPNP_BOT_TOKEN="123456:ABC-DEF1234..." \
    -e TGBOT_UPNP_ADMIN_ID="123456789" \
    -e TGBOT_UPNP_HTTP_PORT=8080 \
    -v /宿主机目录:/data \
    --network host \
    tgbotupnp/tgbot-upnp:latest
```

使用 `--network host` 可自动发现局域网 UPnP 设备。如使用反向代理，设置 `TGBOT_UPNP_BASE_URL`。

镜像同时发布在 Docker Hub 和 GitHub Container Registry：

```bash
# Docker Hub
docker pull tgbotupnp/tgbot-upnp:latest

# GitHub Container Registry
docker pull ghcr.io/tgbot-upnp/tgbot-upnp:latest
```

### 环境变量

| 变量 | 说明 |
|------|------|
| `TGBOT_UPNP_APP_ID` | API ID（使用 2040 即内置 Telegram Desktop） |
| `TGBOT_UPNP_API_HASH` | API Hash |
| `TGBOT_UPNP_BOT_TOKEN` | 从 @BotFather 获取的 Bot Token |
| `TGBOT_UPNP_ADMIN_ID` | 管理员用户 ID，多个用英文逗号分隔（开启自动管理员时可不填） |
| `TGBOT_UPNP_AUTO_ADMIN` | 扫码登录用户自动设为管理员（默认 false） |
| `TGBOT_UPNP_HTTP_PORT` | HTTP 服务端口（默认 8080） |
| `TGBOT_UPNP_BASE_URL` | 自定义播放地址前缀（可选） |
| `TGBOT_UPNP_USER_SESSION` | 私有频道访问用户 Session（扫码登录后自动填充） |
| `TGBOT_UPNP_DATA_DIR` | 数据和配置存储目录（默认当前目录） |

## 使用说明

1. 在 Telegram 中向 Bot 发送视频
2. 点击 ▶️ 播放按钮，选择投屏设备
3. 也可以发送消息链接：`https://t.me/频道名/12345`（支持公开和私有频道）
4. 私有频道需要先在配置页面扫码登录

系统托盘图标（Windows / macOS）提供使用说明、配置页面和退出功能。

## 已测试环境

- [x] Windows 10/11 amd64
- [x] macOS amd64、arm64 (Apple Silicon)
- [x] Linux amd64、arm64
- [x] Docker amd64、arm64、arm/v7、arm/v5

## 已测试投屏软件

- [BubbleUPnP](https://play.google.com/store/apps/details?id=com.bubblesoft.android.bubbleupnp)
- [当贝投屏](https://www.dangbei.com/app/tv/2021/1214/7921.html)
