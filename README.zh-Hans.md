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
- 系统托盘菜单（Windows）
- 支持消息链接投屏（`t.me/频道名/消息ID`）
- 支持 Docker 部署，数据持久化

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
admin_id: 123456789                   # 多个用英文逗号分隔: 123456,789012
http_port: 8080
base_url: ""                          # 可选，反向代理地址如 http://your-proxy.com:8080
```

### Docker 运行

```shell
docker run -d --name tgbot-upnp \
    -e TELEGRAM_APP_ID="2040" \
    -e TELEGRAM_API_HASH="b18441a1ff607e10a989891a5462e627" \
    -e TELEGRAM_BOT_TOKEN="123456:ABC-DEF1234..." \
    -e TELEGRAM_ADMIN_ID="123456789" \
    -e TELEGRAM_HTTP_PORT=8080 \
    -v /宿主机目录:/data \
    --network host \
    tgbotupnp/tgbot-upnp:latest
```

使用 `--network host` 可自动发现局域网 UPnP 设备。如使用反向代理，设置 `TELEGRAM_BASE_URL`。

### 环境变量

| 变量 | 说明 |
|------|------|
| `TELEGRAM_APP_ID` | API ID（使用 2040 即内置 Telegram Desktop） |
| `TELEGRAM_API_HASH` | API Hash |
| `TELEGRAM_BOT_TOKEN` | 从 @BotFather 获取的 Bot Token |
| `TELEGRAM_ADMIN_ID` | 管理员用户 ID，多个用英文逗号分隔 |
| `TELEGRAM_HTTP_PORT` | HTTP 服务端口（默认 8080） |
| `TELEGRAM_BASE_URL` | 自定义播放地址前缀（可选） |
| `TELEGRAM_DATA_DIR` | 数据和配置存储目录（默认当前目录） |

## 使用说明

1. 在 Telegram 中向 Bot 发送视频
2. 点击 ▶️ 播放按钮，选择投屏设备
3. 也可以发送消息链接：`https://t.me/频道名/12345`

系统托盘图标提供使用说明、配置页面和退出功能。

## 已测试环境

- [x] Windows 10/11 amd64
- [x] Docker amd64、arm64、arm/v7、arm/v5
- [x] Linux amd64、arm64

## 已测试投屏软件

- [BubbleUPnP](https://play.google.com/store/apps/details?id=com.bubblesoft.android.bubbleupnp)
- [当贝投屏](https://www.dangbei.com/app/tv/2021/1214/7921.html)
