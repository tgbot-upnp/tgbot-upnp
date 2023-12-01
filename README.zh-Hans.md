# tgbot-upnp
tgbot-upnp 是一个可以将 telegram 视频通过 upnp 协议投屏至其他设备的 telegram 机器人

[English](https://github.com/tgbot-upnp/tgbot-upnp/blob/master/README.md) | [简体中文](https://github.com/tgbot-upnp/tgbot-upnp/blob/master/README.zh-Hans.md)
## 特性

- 将 telegram 视频投屏到支持 UPnP (AVTransport) 的设备
- 无需下载整个视频，支持进度条
- 自动扫描局域网中所有支持 UPnP (AVTransport) 的设备
- 支持多语言
- 较低的内存和cpu占用
- 支持 Windows、docker运行

## 快速开始
### 准备工作
1. 创建自己的机器人 https://telegram.me/BotFather

   获取`API Token`

2. 创建自己的应用 https://core.telegram.org/api/obtaining_api_id
   
   获取 `App api_id` `App api_hash`

   telegram官方bot的api不支持串流下载，实现串流只能使用支持MTProto2.0协议的第三方客户端,使用第三方客户端需要申请自己的api_id

3. 获取自己的用户ID https://telegram.me/userinfobot
   
   获取 `userID` 整数型数字
   
   用于机器人判断当前用户是否能为管理员
### Windows 运行

从release页面下载对应的安装包，编辑config.yml文件，替换为准备工作获取到的信息
```yaml
app_id: App api_id
api_hash: App api_hash
bot_token: API Toke
http_port: 8080
admin_id: userID #多个用户ID用“,”分隔: userID1,userID2
```
启动应用程序即可，可以在系统托盘看到已经运行的程序


### docker 运行

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

## 测试运行环境
- [x] windows10-amd64
- [x] windows10-386
- [x] Docker-amd64
## 已测试投屏软件

-  [BubbleUPNP](https://play.google.com/store/apps/details?id=com.bubblesoft.android.bubbleupnp)
-  [当贝投屏](https://www.dangbei.com/app/tv/2021/1214/7921.html)

## Todo
- [ ] 支持mac
- [ ] 支持openwrt