package bot

import (
	"context"
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/clock"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tgbot-upnp/tgbot-upnp/dcpool"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"github.com/tgbot-upnp/tgbot-upnp/middleware/retry"
	"github.com/tgbot-upnp/tgbot-upnp/server"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
	"golang.org/x/text/language"
	"slices"
)

const (
	cmdStart         = "start"
	cbPlay           = "Play"
	cbPlayWithDevice = "P-"
	cbCachedVideo    = "CV-"
	cbCachedDevice   = "CD-"
)

var botLogger *zap.Logger
var botAPI *tg.Client
var autoAdminUserID atomic.Int64

func Run(ctx context.Context, appID int, apiHash, botToken, userSession, sessionDir string, adminIDs []int, autoAdmin bool, log *zap.Logger) error {
	botLogger = log
	createUserClient(ctx, appID, apiHash, userSession, autoAdmin)

	sessionPath := filepath.Join(sessionDir, "tgbot-upnp.session")
	sess := newFileSession(sessionPath)

	disp := NewDispatcher(botLogger)
	disp.AddMiddleware(authMiddleware(adminIDs))
	disp.AddCommand(cmdStart, handleStart)
	disp.AddMessageVideo(handleVideo)
	disp.AddMessageText("t.me/", handleTextLink)
	disp.AddCallbackExact(cbPlay, handleCbPlay)
	disp.AddCallbackPrefix(cbPlayWithDevice, handleCbPlayWithDevice)
	disp.AddCallbackPrefix(cbCachedVideo, handleCbCachedVideo)
	disp.AddCallbackPrefix(cbCachedDevice, handleCbCachedDevice)

	client := telegram.NewClient(appID, apiHash, telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{Dial: proxy.Direct.DialContext}),
		ReconnectionBackoff: func() backoff.BackOff {
			b := backoff.NewExponentialBackOff()
			b.Multiplier = 1.1
			b.MaxElapsedTime = 0
			b.MaxInterval = 10 * time.Second
			return b
		},
		Device:         telegram.DeviceConfig{DeviceModel: "tgbot-upnp", SystemVersion: "1.0", AppVersion: "1.0"},
		SessionStorage: sess,
		RetryInterval:  5 * time.Second,
		MaxRetries:     5,
		DialTimeout:    10 * time.Second,
		Middlewares:    []telegram.Middleware{floodwait.NewSimpleWaiter()},
		Clock:          clock.System,
		UpdateHandler: telegram.UpdateHandlerFunc(
			func(ictx context.Context, u tg.UpdatesClass) error {
				for _, u2 := range extractUpdates(u) {
					logUpdate(&u2)
					disp.Dispatch(newBotContext(ictx, botAPI, nil), &u2)
				}
				return nil
			},
		),
	})
	botAPI = client.API()

	return client.Run(ctx, func(ctx context.Context) error {
		if err := authBot(ctx, client, botToken); err != nil {
			return fmt.Errorf("bot auth: %w", err)
		}
		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("resolve self: %w", err)
		}
		botLogger.Info("bot authorized", zap.String("username", self.Username))

		pool := dcpool.New(client, 4, botLogger, retry.New(5, botLogger))
		server.SetPool(pool)
		botLogger.Info("dc pool initialized")
		setBotInfo(client.API())
		setBotCommands(client.API())

		<-ctx.Done()
		return ctx.Err()
	})
}

func authBot(ctx context.Context, client *telegram.Client, token string) error {
	botLogger.Info("checking auth status")
	status, err := client.Auth().Status(ctx)
	if err != nil {
		botLogger.Warn("auth status check failed", zap.Error(err))
		return fmt.Errorf("auth status: %w", err)
	}
	if !status.Authorized {
		botLogger.Info("performing bot login")
		if _, err := client.Auth().Bot(ctx, token); err != nil {
			botLogger.Error("bot login failed", zap.Error(err))
			return fmt.Errorf("bot login: %w", err)
		}
		botLogger.Info("bot login success")
	}
	return nil
}

func setBotInfo(api *tg.Client) {
	ctx := context.Background()
	localizer := lang.GetLocalizer(lang.DefaultTag)
	_, _ = api.BotsSetBotInfo(ctx, &tg.BotsSetBotInfoRequest{
		About: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "TgBotAbout", Other: "cast telegram videos to other devices through the upnp protocol."},
		}),
		Description: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "TgBotDescription", Other: "tgbot-upnp casts telegram videos to upnp devices."},
		}),
	})
	for _, tag := range lang.GetAllSupportedTag() {
		base, _ := tag.Base()
		loc := lang.GetLocalizer(tag)
		_, _ = api.BotsSetBotInfo(ctx, &tg.BotsSetBotInfoRequest{
			LangCode: base.String(),
			About: loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "TgBotAbout", Other: "cast telegram videos to other devices."},
			}),
			Description: loc.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "TgBotDescription", Other: "tgbot-upnp is a small tool for upnp casting."},
			}),
		})
	}
	botLogger.Info("bot info updated")
}

func setBotCommands(api *tg.Client) {
	ctx := context.Background()
	addCmd := func(langCode string, loc *i18n.Localizer) {
		_, err := api.BotsSetBotCommands(ctx, &tg.BotsSetBotCommandsRequest{
			Scope:    &tg.BotCommandScopeDefault{},
			LangCode: langCode,
			Commands: []tg.BotCommand{
				{Command: "start", Description: loc.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{ID: "TgBotCmdStartDesc", Other: "Show usage instructions"},
				})},
			},
		})
		if err != nil {
			botLogger.Warn("set bot commands failed", zap.String("lang", langCode), zap.Error(err))
		}
	}
	addCmd("", lang.GetLocalizer(lang.DefaultTag))
	addCmd("en", lang.GetLocalizer(lang.DefaultTag))
	addCmd("zh", lang.GetLocalizer(language.Make("zh-Hans")))
	botLogger.Info("bot commands updated")
}

func authMiddleware(adminIDs []int) MiddlewareFunc {
	return func(ctx *Context, u *Update) error {
		chatID := getChatID(u)
		if chatID == 0 {
			botLogger.Info("update from non-user chat, ignoring")
			return ErrStopDispatch
		}
		if !slices.Contains(adminIDs, int(chatID)) && int64(chatID) != autoAdminUserID.Load() {
			botLogger.Info("non-admin access", zap.Int64("userID", chatID))
			localizer := lang.GetLocalizer(userLang(u))
			ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{
				Message: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "TgBotAuthFailed",
						Other: "Your user ID: {{.userID}}\nYou are not the administrator of the current tgbot-upnp and cannot use this bot.",
					},
					TemplateData: map[string]interface{}{"userID": chatID},
				}),
			})
			return ErrStopDispatch
		}
		return nil
	}
}

func logUpdate(u *Update) {
	if u.Message != nil {
		fields := []zap.Field{
			zap.Int64("userID", u.UserID),
			zap.String("lang", u.UserLang),
			zap.Int("msgID", u.Message.ID),
		}
		if u.Message.Message != "" {
			fields = append(fields, zap.String("text", u.Message.Message))
		}
		if media, ok := u.Message.GetMedia(); ok {
			switch m := media.(type) {
			case *tg.MessageMediaDocument:
				if doc, ok := m.Document.(*tg.Document); ok {
					fields = append(fields,
						zap.String("type", "document"),
						zap.String("mime", doc.MimeType),
						zap.Int64("size", doc.Size),
						zap.String("fileName", getDocumentName(doc)),
					)
					for _, attr := range doc.Attributes {
						if v, ok := attr.(*tg.DocumentAttributeVideo); ok {
							fields = append(fields,
								zap.String("type", "video"),
								zap.Float64("duration", v.Duration),
								zap.Int("width", v.W),
								zap.Int("height", v.H),
							)
						}
					}
				}
			case *tg.MessageMediaPhoto:
				if photo, ok := m.Photo.(*tg.Photo); ok {
					fields = append(fields,
						zap.String("type", "photo"),
						zap.Int64("id", photo.ID),
					)
				}
			default:
				fields = append(fields, zap.String("type", "other"))
			}
		}
		botLogger.Info("received message", fields...)
	} else if u.CallbackQuery != nil {
		botLogger.Info("received callback",
			zap.Int64("userID", u.UserID),
			zap.String("lang", u.UserLang),
			zap.String("prefix", string(u.CallbackQuery.Data[:min(1, len(u.CallbackQuery.Data))])),
			zap.Int("len", len(u.CallbackQuery.Data)),
		)
	}
}

func getDocumentName(doc *tg.Document) string {
	for _, attr := range doc.Attributes {
		if f, ok := attr.(*tg.DocumentAttributeFilename); ok {
			return f.FileName
		}
	}
	return ""
}

func getChatID(u *Update) int64 {
	if u.Message != nil {
		switch p := u.Message.PeerID.(type) {
		case *tg.PeerUser:
			return p.UserID
		case *tg.PeerChat:
			return p.ChatID
		}
	}
	if u.CallbackQuery != nil {
		return u.CallbackQuery.UserID
	}
	return 0
}

func userLang(u *Update) language.Tag {
	if u.UserLang != "" {
		return language.Make(u.UserLang)
	}
	return language.English
}
