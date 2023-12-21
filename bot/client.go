package bot

import (
	"bytes"
	"crypto/md5"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/huin/goupnp/dcps/av1"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tgbot-upnp/tgbot-upnp/http"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"github.com/tgbot-upnp/tgbot-upnp/upnp"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"golang.org/x/time/rate"
	"slices"
	"time"
)

const (
	tBotCmdStart         = "start"
	tBotCbPlay           = "Play"
	tBotCbPlayWithDevice = "P-"
)

type groupData struct {
	Message   string
	GroupedID int64
}

var logger *zap.Logger
var cacheGroupData = groupData{}

func Client(appId int, apiHash, botToken string, adminIDs []int, globalLogger *zap.Logger) {
	logger = globalLogger
	client, err := gotgproto.NewClient(
		appId,
		apiHash,
		gotgproto.ClientType{
			BotToken: botToken,
		},
		&gotgproto.ClientOpts{
			Logger:           logger,
			Session:          sessionMaker.SqliteSession("tgbot-upnp"),
			DisableCopyright: true,
			Middlewares: []telegram.Middleware{
				ratelimit.New(rate.Every(500*time.Millisecond), 1),
			},
		},
	)
	if err != nil {
		logger.Fatal("tgbot-upnp tg client failed to start :", zap.String("err", err.Error()))
	}
	tBotInit(client)
	client.Dispatcher.AddHandler(handlers.NewAnyUpdate(func(ctx *ext.Context, update *ext.Update) error {
		return baseAuth(adminIDs, ctx, update)
	}))
	client.Dispatcher.AddHandler(handlers.NewCommand(tBotCmdStart, cmdStart))
	client.Dispatcher.AddHandler(handlers.NewMessage(filters.Message.Video, msgVideo))
	client.Dispatcher.AddHandler(handlers.NewCallbackQuery(filters.CallbackQuery.Equal(tBotCbPlay), cbPlay))
	client.Dispatcher.AddHandler(handlers.NewCallbackQuery(filters.CallbackQuery.Prefix(tBotCbPlayWithDevice), cbPlayWithDevice))
	logger.Info("tgbot-upnp tg client has been started...", zap.String("bot", client.Self.Username))
	err = client.Idle()
	if err != nil {
		logger.Fatal("tgbot-upnp tg client failed on idle :", zap.String("err", err.Error()))
	}
}
func tBotInit(client *gotgproto.Client) {
	localizer := lang.GetLocalizer(lang.DefaultTag)
	_, _ = client.API().BotsSetBotInfo(client.CreateContext(), &tg.BotsSetBotInfoRequest{
		About: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "TgBotAbout",
				Other: "cast telegram videos to other devices through the upnp protocol. https://github.com/tgbot-upnp/tgbot-upnp",
			},
		}),
		Description: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "TgBotDescription",
				Other: "tgbot-upnp is a small tool that can cast telegram videos to other devices through the upnp protocol. You can send the video to the current conversation to start your screen casting experience. details: https://github.com/tgbot-upnp/tgbot-upnp",
			},
		}),
	})
	tags := lang.GetAllSupportedTag()
	for _, tag := range tags {
		localizer := lang.GetLocalizer(tag)
		tagBase, _ := tag.Base()
		_, err := client.API().BotsSetBotInfo(client.CreateContext(), &tg.BotsSetBotInfoRequest{
			LangCode: tagBase.String(),
			About: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "TgBotAbout",
					Other: "tgbot-upnp is a small tool that can cast telegram videos to other devices through the upnp protocol\n https://github.com/tgbot-upnp/tgbot-upnp",
				},
			}),
			Description: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "TgBotDescription",
					Other: "tgbot-upnp is a small tool that can cast telegram videos to other devices through the upnp protocol. You can send the video to the current conversation to start your screen casting experience. details: https://github.com/tgbot-upnp/tgbot-upnp",
				},
			}),
		})
		if err != nil {
			logger.Error("BotsSetBotInfo", zap.String("err", err.Error()))
		}
	}
	logger.Info("Init bot command succeed")
}
func baseAuth(adminIDs []int, ctx *ext.Context, update *ext.Update) error {
	if update.EffectiveChat().IsAUser() {
		if !slices.Contains(adminIDs, int(update.GetUserChat().GetID())) {
			logger.Info("Non-administrative user access", zap.String("username", update.GetUserChat().Username), zap.Int64("userID", update.GetUserChat().GetID()))
			localizer := lang.GetLocalizer(language.Make(update.GetUserChat().LangCode))
			_, _ = ctx.SendMessage(update.GetUserChat().GetID(), &tg.MessagesSendMessageRequest{Message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "TgBotAuthFailed",
					Other: "Your user ID: {{.userID}}\nYou are not the administrator of the current tgbot-upnp and cannot use this bot. You can contact the administrator to add your user ID to the administrator list, or deploy your own tgbot-upnp, details: https://github.com/tgbot-upnp/tgbot-upnp",
				},
				TemplateData: map[string]interface{}{
					"userID": update.GetUserChat().GetID(),
				},
			})})
			return dispatcher.EndGroups
		}
	} else {
		logger.Info("tgbot-upnp is only work in user chat")
		return dispatcher.EndGroups
	}
	return nil
}
func cmdStart(ctx *ext.Context, update *ext.Update) error {
	localizer := lang.GetLocalizer(language.Make(update.GetUserChat().LangCode))
	_, _ = ctx.SendMessage(update.GetUserChat().GetID(), &tg.MessagesSendMessageRequest{Message: localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "TgBotCmdStart",
			Other: "Welcome to tgbot-upnp, you can send videos to the current conversation to start your screencasting experience",
		},
		TemplateData: map[string]interface{}{
			"userID": update.GetUserChat().GetID(),
		},
	})})
	return nil
}
func getDevicesReplyInlineMarkup(update *ext.Update) *tg.ReplyInlineMarkup {
	av1Clients, _, _ := av1.NewAVTransport1Clients()
	rows := make([]tg.KeyboardButtonRow, 0)
	for _, av1Client := range av1Clients {
		av1ClientMd5 := md5.Sum([]byte(av1Client.ServiceClient.RootDevice.URLBase.Host))
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: "▶️ " + av1Client.ServiceClient.RootDevice.Device.FriendlyName,
					Data: append([]byte(tBotCbPlayWithDevice), av1ClientMd5[:]...),
				},
			}})
	}
	localizer := lang.GetLocalizer(language.Make(update.GetUserChat().LangCode))
	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonCallback{
				Text: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "TgBotMsgRefresh",
						Other: "🔄 Refresh",
					},
				}),
				Data: []byte(tBotCbPlay),
			},
		}})
	return &tg.ReplyInlineMarkup{
		Rows: rows,
	}
}
func msgVideo(ctx *ext.Context, update *ext.Update) error {
	logger.Info("sourceMessage", zap.Any("message", update.EffectiveMessage))
	localizer := lang.GetLocalizer(language.Make(update.GetUserChat().LangCode))
	sourceMessage := update.EffectiveMessage.Message.Message
	if groupedID, ok := update.EffectiveMessage.Message.GetGroupedID(); ok {
		if cacheGroupData.GroupedID == groupedID {
			sourceMessage = cacheGroupData.Message
		} else {
			cacheGroupData.GroupedID = groupedID
			cacheGroupData.Message = sourceMessage
		}
	}
	msgVideoInput := convInputMedia(update.EffectiveMessage.Media)
	media, err := ctx.SendMedia(update.GetUserChat().GetID(), &tg.MessagesSendMediaRequest{
		Message: sourceMessage,
		Media:   msgVideoInput,
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonCallback{
						Text: localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{
								ID:    "TgBotMsgPlay",
								Other: "▶️ Play",
							},
						}),
						Data: []byte(tBotCbPlay),
					},
				}}},
		},
	})
	if err != nil {
		logger.Error("tgbot-upnp reply video msg error", zap.String("err", err.Error()))
		return err
	}
	logger.Info("tgbot-upnp reply video msg", zap.Any("media", media))
	return nil
}
func cbPlayWithDevice(ctx *ext.Context, update *ext.Update) error {
	av1Clients, _, _ := av1.NewAVTransport1Clients()
	localizer := lang.GetLocalizer(language.Make(update.GetUserChat().LangCode))
	if len(av1Clients) > 0 {
		for _, av1Client := range av1Clients {
			av1ClientMd5 := md5.Sum([]byte(av1Client.ServiceClient.RootDevice.URLBase.Host))
			if bytes.Equal(update.CallbackQuery.Data, append([]byte(tBotCbPlayWithDevice), av1ClientMd5[:]...)) {
				err := videoPlay(av1Client, ctx, update)
				if err != nil {
					logger.Error("Play video error", zap.String("err", err.Error()))
				}
				_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
					Alert:   false,
					QueryID: update.CallbackQuery.QueryID,
					Message: localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "TgBotMsgVideoPlayed",
							Other: "Video has started playing",
						},
					}),
				})
				return nil
			}
		}
	}
	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		Alert:   true,
		QueryID: update.CallbackQuery.QueryID,
		Message: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "TgBotMsgDeviceUnavailable",
				Other: "The playback device is unavailable, please refresh first",
			},
		}),
	})
	return nil
}
func cbPlay(ctx *ext.Context, update *ext.Update) error {
	_, _ = ctx.EditMessage(update.GetUserChat().GetID(), &tg.MessagesEditMessageRequest{
		ID:          update.CallbackQuery.MsgID,
		ReplyMarkup: getDevicesReplyInlineMarkup(update),
	})
	_, _ = ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: update.CallbackQuery.QueryID,
	})
	return nil
}
func videoPlay(avClient *av1.AVTransport1, ctx *ext.Context, update *ext.Update) error {
	sourceMsg, _ := ctx.GetMessages(update.GetUserChat().GetID(), []tg.InputMessageClass{&tg.InputMessageID{
		ID: update.CallbackQuery.MsgID,
	}})
	MessageMediaDocument := sourceMsg[0].(*tg.Message).Media.(*tg.MessageMediaDocument).Document.(*tg.Document)
	videoPlayUrl, tgVideoID, _ := http.GetTgVideoPlayUrl(&http.TgVideo{
		Api: ctx.Raw,
		Doc: MessageMediaDocument,
	}, avClient.LocalAddr())
	metadata := upnp.GetMetaData(sourceMsg[0].(*tg.Message).Message, tgVideoID, MessageMediaDocument)
	logger.Info("Play tg Video", zap.String("url", videoPlayUrl))
	_ = avClient.Stop(0)
	_ = avClient.SetAVTransportURI(0, videoPlayUrl, metadata)
	_ = avClient.Play(0, "1")
	return nil
}
func convInputMedia(media tg.MessageMediaClass) tg.InputMediaClass {
	MessageMediaDocument := media.(*tg.MessageMediaDocument)
	Document := MessageMediaDocument.Document.(*tg.Document)
	InputDocument := &tg.InputDocument{}
	InputDocument.FillFrom(Document)
	InputMediaDocument := &tg.InputMediaDocument{
		Spoiler:    MessageMediaDocument.Spoiler,
		ID:         InputDocument,
		TTLSeconds: MessageMediaDocument.TTLSeconds,
	}
	InputMediaDocument.SetFlags()
	return InputMediaDocument
}
