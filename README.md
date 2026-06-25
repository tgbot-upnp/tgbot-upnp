# tgbot-upnp

tgbot-upnp is a Telegram bot that casts videos to UPnP-enabled devices via streaming — no need to download the entire file.

[English](https://github.com/tgbot-upnp/tgbot-upnp/blob/main/README.md) | [简体中文](https://github.com/tgbot-upnp/tgbot-upnp/blob/main/README.zh-Hans.md)

## Features

- Stream Telegram videos to UPnP (AVTransport) devices
- No full download required — streaming playback with progress
- Auto-discover all UPnP devices on LAN
- Multi-language support (English, 简体中文)
- Low memory and CPU usage
- Built-in credential presets — no Telegram API registration needed
- Browser-based config page for easy setup
- System tray menu (Windows / macOS)
- QR code login for private channel access
- Support video links (`t.me/channel/msgID`), including private channels
- Auto-admin: QR-scanned user automatically granted admin access
- Docker support with data persistence and multi-arch images

## Quick Start

### Option 1: Built-in Credentials (Recommended)

1. Create a bot at [@BotFather](https://t.me/BotFather) → get `API Token`
2. Get your user ID from [@userinfobot](https://t.me/userinfobot)
3. Download from [releases](https://github.com/tgbot-upnp/tgbot-upnp/releases), run the app
4. A browser-based setup page opens automatically — select "Telegram Desktop (official)" for built-in API credentials, fill in your bot token and admin ID, save

### Option 2: Windows (manual)

Edit `config.yml`:

```yaml
app_id: 2040                          # or your own from my.telegram.org/apps
api_hash: b18441a1ff607e10a989891a5462e627
bot_token: 123456:ABC-DEF1234...
admin_id: 123456789                   # or 123456,789012 for multiple; optional if auto_admin is true
auto_admin: true                      # auto-set QR-scanned user as admin
http_port: 8080
base_url: ""                          # optional: http://your-proxy.com:8080
user_session: ""                      # auto-filled by QR login
```

### macOS

Download the `.app` bundle from [releases](https://github.com/tgbot-upnp/tgbot-upnp/releases), move to `/Applications`.

First launch — right-click `tgbot-upnp.app` → Open → confirm. This is only needed once (Gatekeeper requires it for unsigned apps).

Config and session files are stored in `~/Library/Application Support/tgbot-upnp/`.

### Docker

```shell
docker run -d --name tgbot-upnp \
    -e TGBOT_UPNP_APP_ID="2040" \
    -e TGBOT_UPNP_API_HASH="b18441a1ff607e10a989891a5462e627" \
    -e TGBOT_UPNP_BOT_TOKEN="123456:ABC-DEF1234..." \
    -e TGBOT_UPNP_ADMIN_ID="123456789" \
    -e TGBOT_UPNP_HTTP_PORT=8080 \
    -v /host/data:/data \
    --network host \
    tgbotupnp/tgbot-upnp:latest
```

With `--network host`, UPnP discovery works automatically on LAN. For reverse proxy setups, set `TGBOT_UPNP_BASE_URL`.

Images are available on both registries:

```bash
# Docker Hub
docker pull tgbotupnp/tgbot-upnp:latest

# GitHub Container Registry
docker pull ghcr.io/tgbot-upnp/tgbot-upnp:latest
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `TGBOT_UPNP_APP_ID` | API ID (use 2040 for built-in Telegram Desktop) |
| `TGBOT_UPNP_API_HASH` | API hash |
| `TGBOT_UPNP_BOT_TOKEN` | Bot token from @BotFather |
| `TGBOT_UPNP_ADMIN_ID` | Admin user ID(s), comma-separated (optional if `AUTO_ADMIN` is set) |
| `TGBOT_UPNP_AUTO_ADMIN` | Auto-set QR-scanned user as admin (default: false) |
| `TGBOT_UPNP_HTTP_PORT` | HTTP server port (default: 8080) |
| `TGBOT_UPNP_BASE_URL` | Custom base URL for reverse proxy (optional) |
| `TGBOT_UPNP_USER_SESSION` | User session for private channel access (auto-filled by QR login) |
| `TGBOT_UPNP_DATA_DIR` | Data directory for config and session (default: `.`) |

## Usage

1. Send a video to the bot in Telegram
2. Click ▶️ Play and select your UPnP device
3. Or send a message link: `https://t.me/channel_name/12345` (public or private channels)
4. For private channels, scan the QR code in the Settings page first to log in

The system tray icon (Windows / macOS) provides quick access to usage help, config page, and quit.

## Tested Environments

- [x] Windows 10/11 amd64
- [x] macOS amd64, arm64 (Apple Silicon)
- [x] Linux amd64, arm64
- [x] Docker amd64, arm64, arm/v7, arm/v5

## Tested UPnP Devices

- [BubbleUPnP](https://play.google.com/store/apps/details?id=com.bubblesoft.android.bubbleupnp)
- [当贝投屏](https://www.dangbei.com/app/tv/2021/1214/7921.html)
