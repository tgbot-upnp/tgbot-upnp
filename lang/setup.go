package lang

import "github.com/nicksnyder/go-i18n/v2/i18n"

// SetupStrings holds all translatable strings for the setup wizard page.
type SetupStrings struct {
	Title              string
	Sub                string
	CredSource         string
	CredBuiltinDesktop string
	CredBuiltinTDL     string
	CredCustom         string
	AppID              string
	AppIDHint          string
	APIHash            string
	APIHashHint        string
	BotToken           string
	BotTokenHint       string
	AdminID            string
	AdminIDHint        string
	HTTPPort           string
	BaseURL            string
	BaseURLHint        string
	UserSession        string
	UserSessionHint    string
	AutoAdmin          string
	BtnScan            string
	QRTitle            string
	QRWaiting          string
	QROK               string
	BtnSave            string
	Saved              string
}

func GetSetupStrings() SetupStrings {
	localizer := GetLocalizer(LocaleSystemTag)
	return SetupStrings{
		Title: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupTitle", Other: "⚙ tgbot-upnp Setup"},
		}),
		Sub: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupSub", Other: "First run detected — please fill in your Telegram credentials."},
		}),
		CredSource: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupCredSource", Other: "App Credentials (api_id & api_hash)"},
		}),
		CredBuiltinDesktop: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupCredBuiltinDesktop", Other: "Telegram Desktop (official)"},
		}),
		CredBuiltinTDL: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupCredBuiltinTDL", Other: "tdl (community)"},
		}),
		CredCustom: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupCredCustom", Other: "自定义"},
		}),
		AppID: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAppID", Other: "App api_id"},
		}),
		AppIDHint: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAppIDHint", Other: "Get it from my.telegram.org/apps"},
		}),
		APIHash: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAPIHash", Other: "App api_hash"},
		}),
		APIHashHint: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAPIHashHint", Other: "Same page as above"},
		}),
		BotToken: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupBotToken", Other: "Bot Token"},
		}),
		BotTokenHint: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupBotTokenHint", Other: "Create one at @BotFather"},
		}),
		AdminID: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAdminID", Other: "Admin User ID(s)"},
		}),
		AdminIDHint: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAdminIDHint", Other: "Get yours at @userinfobot. Use commas to separate multiple IDs like 123,456."},
		}),
		HTTPPort: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupHTTPPort", Other: "HTTP Port"},
		}),
		BaseURL: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupBaseURL", Other: "Base URL (optional)"},
		}),
		BaseURLHint: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupBaseURLHint", Other: "Reverse proxy address. Leave empty to auto-detect."},
		}),
		UserSession: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupUserSession", Other: "User Session (optional)"},
		}),
		UserSessionHint: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupUserSessionHint", Other: "Login with QR code to access private channel videos."},
		}),
		AutoAdmin: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupAutoAdmin", Other: "Auto-set logged-in user as admin"},
		}),
		BtnScan: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupBtnScan", Other: "📱 Scan QR"},
		}),
		QRTitle: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupQRTitle", Other: "Scan QR Code"},
		}),
		QRWaiting: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupQRWaiting", Other: "Open Telegram on your phone → Settings → Devices → Scan QR Code"},
		}),
		QROK: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupQROK", Other: "✅ Logged in!"},
		}),
		BtnSave: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupBtnSave", Other: "💾 Save"},
		}),
		Saved: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "SetupSaved", Other: "✅ Configuration saved!"},
		}),
	}
}
