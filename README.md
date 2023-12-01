# tgbot-upnp
tgbot-upnp is a telegram bot that can cast telegram videos to other devices via upnp protocol



[English](https://github.com/tgbot-upnp/tgbot-upnp/blob/main/README.md) | [简体中文](https://github.com/tgbot-upnp/tgbot-upnp/blob/main/README.zh-Hans.md)

This description was generated using translation software. Help with translation is welcome.

## Feature

- Cast telegram videos to devices that support UPnP (AVTransport)
- No need to download the entire video, supports progress bar
- Automatically scan all UPnP (AVTransport)-enabled devices on the LAN
- Support multiple languages
- Lower memory and cpu usage
- Support windows/docker running

## Quick start
### Preparation
1. Create your own bot https://telegram.me/BotFather

   GET `API Token`

2. Create your own app https://core.telegram.org/api/obtaining_api_id

   GET `App api_id` `App api_hash`

   The API of Telegram’s official bot does not support streaming downloads. To achieve streaming, you can only use a third-party client that supports the MTProto2.0 protocol. To use a third-party client, you need to apply for your own api_id.

3. Get your user ID https://telegram.me/userinfobot

   GET `userID` integer number

   Used by bot to determine whether the current user can be an administrator
### windows

Download the app from [release](https://github.com/tgbot-upnp/tgbot-upnp/releases) page , edit the config.yml file, and replace it with the information obtained during the preparation work.
```yaml
app_id: App api_id
api_hash: App api_hash
bot_token: API Toke
http_port: 8080
admin_id: userID #Multiple user IDs are separated by ",": userID1,userID2
```
Just start the application and you can see the running program in the system tray


### docker

```shell
docker run -d --name tgbot-upnp \
    -e TELEGRAM_APP_ID="App api_id" \
    -e TELEGRAM_API_HASH="App api_hash" \
    -e TELEGRAM_BOT_TOKEN="API Toke" \
    -e TELEGRAM_HTTP_PORT=8080 \
    -e TELEGRAM_ADMIN_ID="userID,userID2" \
    --network host \
    tgbotupnp/tgbot-upnp:latest
```

## Test run environment
- [x] Windows10-amd64
- [x] Windows10-386
- [x] Docker-amd64
## Screen-casting software tested

-  [BubbleUPNP](https://play.google.com/store/apps/details?id=com.bubblesoft.android.bubbleupnp)
-  [当贝投屏](https://www.dangbei.com/app/tv/2021/1214/7921.html)

## Todo
- [ ] support mac
- [ ] support openwrt