package main

import (
	"github.com/spf13/viper"
	"github.com/tgbot-upnp/tgbot-upnp/bot"
	"github.com/tgbot-upnp/tgbot-upnp/config"
	"github.com/tgbot-upnp/tgbot-upnp/http"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"go.uber.org/zap"
)

var logger *zap.Logger

func main() {
	go bot.Client(viper.GetInt(config.AppID), viper.GetString(config.ApiHash), viper.GetString(config.BotToken), viper.GetIntSlice(config.AdminIDs), logger)
	http.Server(viper.GetInt(config.HttpPort), logger)
	select {}
}

func init() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	config.GetConfig(logger)
	lang.GetI18nBundle(logger)
}
