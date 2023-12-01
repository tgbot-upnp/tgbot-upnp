package main

import (
	"github.com/getlantern/systray"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/viper"
	"github.com/tgbot-upnp/tgbot-upnp/bot"
	"github.com/tgbot-upnp/tgbot-upnp/config"
	"github.com/tgbot-upnp/tgbot-upnp/http"
	"github.com/tgbot-upnp/tgbot-upnp/icon"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"go.uber.org/zap"
	"os"
)

var logger *zap.Logger

func main() {
	go bot.Client(viper.GetInt(config.AppID), viper.GetString(config.ApiHash), viper.GetString(config.BotToken), viper.GetIntSlice(config.AdminIDs), logger)
	http.Server(viper.GetInt(config.HttpPort), logger)
	tray()
	select {}
}

func init() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	config.GetConfig(logger)
	lang.GetI18nBundle(logger)
}

func tray() {
	onReady := func() {
		localizer := lang.GetLocalizer(lang.LocaleSystemTag)
		systray.SetIcon(icon.GetIcon())
		systray.SetTitle("tgbot-upnp")
		systray.SetTooltip("tgbot-upnp")
		mQuit := systray.AddMenuItem(localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SystrayQuit",
				Other: "Quit",
			},
		}), localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SystrayQuitTooltip",
				Other: "Quit tgbot-upnp",
			},
		}))
		for range mQuit.ClickedCh {
			os.Exit(0)
		}
	}
	onExit := func() {}
	systray.Run(onReady, onExit)
}
