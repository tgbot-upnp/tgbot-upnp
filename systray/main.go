//go:build windows

package main

import (
	"fmt"

	"github.com/getlantern/systray"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tgbot-upnp/tgbot-upnp/app"
	"github.com/tgbot-upnp/tgbot-upnp/icon"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"github.com/tgbot-upnp/tgbot-upnp/setup"
)

var trayApp *app.App

func main() {
	// Start tray immediately so it's visible during first-run setup
	go tray()

	a := app.New()
	trayApp = a // make available to tray menu handlers
	defer a.Wait()
	select {}
}

func tray() {
	a := trayApp // may be nil initially, handlers check this
	onReady := func() {
		localizer := lang.GetLocalizer(lang.LocaleSystemTag)
		systray.SetIcon(icon.GetIcon())
		systray.SetTitle("tgbot-upnp")
		systray.SetTooltip("tgbot-upnp")

		usageText := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayUsage", Other: "Usage"},
		})
		usageTitle := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayUsageTitle", Other: "tgbot-upnp"},
		})
		usageMsg := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayUsageMsg", Other: "Send a video to this bot in Telegram.\nThen click the Play button and select your UPnP device."},
		})
		mUsage := systray.AddMenuItem(usageText, "")
		go func() {
			for range mUsage.ClickedCh {
				showMessage(usageTitle, usageMsg)
			}
		}()

		configText := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayOpenConfig", Other: "Open Config"},
		})
		mConfig := systray.AddMenuItem(configText, "")
		go func() {
			for range mConfig.ClickedCh {
				port := 8080
				if a != nil {
					port = a.HTTPPort
				}
				_ = setup.OpenBrowser(fmt.Sprintf("http://127.0.0.1:%d", port))
			}
		}()

		autoText := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayAutostart", Other: "Start with Windows"},
		})
		mAuto := systray.AddMenuItemCheckbox(autoText, "", isAutostartEnabled())
		go func() {
			for range mAuto.ClickedCh {
				if mAuto.Checked() {
					_ = setAutostart(false)
					mAuto.Uncheck()
				} else {
					_ = setAutostart(true)
					mAuto.Check()
				}
			}
		}()

		systray.AddSeparator()

		quitText := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayQuit", Other: "Quit"},
		})
		quitTip := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SystrayQuitTooltip", Other: "Quit tgbot-upnp"},
		})
		mQuit := systray.AddMenuItem(quitText, quitTip)
		go func() {
			for range mQuit.ClickedCh {
				a.Cancel()
			}
		}()
	}
	onExit := func() {}
	systray.Run(onReady, onExit)
}
