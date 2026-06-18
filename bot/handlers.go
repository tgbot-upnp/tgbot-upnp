package bot

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/huin/goupnp/dcps/av1"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"github.com/tgbot-upnp/tgbot-upnp/server"
	"github.com/tgbot-upnp/tgbot-upnp/upnp"
	"go.uber.org/zap"
)

// t.me link pattern: optional https://, then t.me/, then username or c/ID, then /msgID
var linkRe = regexp.MustCompile(`(?:https?://)?t\.me/((?:c/)?[a-zA-Z0-9_+-]+)/(\d+)`)

func handleStart(ctx *Context, u *Update) error {
	chatID := getChatID(u)
	localizer := lang.GetLocalizer(userLang(u))
	ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{
		Message: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "TgBotCmdStart",
				Other: "Welcome to tgbot-upnp, you can send videos to the current conversation to start your screencasting experience",
			},
		}),
	})
	return nil
}

func handleVideo(ctx *Context, u *Update) error {
	chatID := getChatID(u)
	localizer := lang.GetLocalizer(userLang(u))

	media, err := ctx.SendMedia(chatID, &tg.MessagesSendMediaRequest{
		Media: convInputMedia(u.Message.Media),
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{{
				Buttons: []tg.KeyboardButtonClass{
					&tg.KeyboardButtonCallback{
						Text: localizer.MustLocalize(&i18n.LocalizeConfig{
							DefaultMessage: &i18n.Message{ID: "TgBotMsgPlay", Other: "▶️ Play"},
						}),
						Data: []byte(cbPlay),
					},
				}},
			},
		},
	})
	if err != nil {
		botLogger.Error("reply video error", zap.Error(err))
		return err
	}
	botLogger.Info("replied with video", zap.Any("media", media))
	return nil
}

func handleCbPlay(ctx *Context, u *Update) error {
	chatID := getChatID(u)
	localizer := lang.GetLocalizer(userLang(u))
	ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
		ID:          u.CallbackQuery.MsgID,
		ReplyMarkup: getDevicesReplyInlineMarkup(localizer),
	})
	ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: u.CallbackQuery.QueryID,
	})
	return nil
}

func handleCbPlayWithDevice(ctx *Context, u *Update) error {
	localizer := lang.GetLocalizer(userLang(u))
	av1Clients, _, _ := av1.NewAVTransport1Clients()
	if len(av1Clients) > 0 {
		for _, av1Client := range av1Clients {
			av1ClientMd5 := md5.Sum([]byte(av1Client.ServiceClient.RootDevice.URLBase.Host))
			if bytes.Equal(u.CallbackQuery.Data, append([]byte(cbPlayWithDevice), av1ClientMd5[:]...)) {
				botLogger.Info("play on device", zap.String("device", av1Client.ServiceClient.RootDevice.Device.FriendlyName), zap.Int("msgID", u.CallbackQuery.MsgID))
				if err := playVideo(ctx, av1Client, u); err != nil {
					botLogger.Error("play video error", zap.Error(err))
				}
				ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
					Alert:   false,
					QueryID: u.CallbackQuery.QueryID,
					Message: localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{ID: "TgBotMsgVideoPlayed", Other: "Video has started playing"},
					}),
				})
				return nil
			}
		}
	}
	botLogger.Warn("no upnp device available")
	ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		Alert:   true,
		QueryID: u.CallbackQuery.QueryID,
		Message: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "TgBotMsgDeviceUnavailable", Other: "The playback device is unavailable, please refresh first"},
		}),
	})
	return nil
}

func getDevicesReplyInlineMarkup(localizer *i18n.Localizer) *tg.ReplyInlineMarkup {
	av1Clients, _, _ := av1.NewAVTransport1Clients()
	names := make([]string, len(av1Clients))
	for i, c := range av1Clients {
		names[i] = c.ServiceClient.RootDevice.Device.FriendlyName
	}
	botLogger.Info("upnp scan", zap.Strings("devices", names))
	rows := make([]tg.KeyboardButtonRow, 0)
	for _, av1Client := range av1Clients {
		av1ClientMd5 := md5.Sum([]byte(av1Client.ServiceClient.RootDevice.URLBase.Host))
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: "▶️ " + av1Client.ServiceClient.RootDevice.Device.FriendlyName,
					Data: append([]byte(cbPlayWithDevice), av1ClientMd5[:]...),
				},
			}})
	}
	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonCallback{
				Text: localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{ID: "TgBotMsgRefresh", Other: "🔄 Refresh"},
				}),
				Data: []byte(cbPlay),
			},
		}})
	return &tg.ReplyInlineMarkup{Rows: rows}
}

func playVideo(ctx *Context, avClient *av1.AVTransport1, u *Update) error {
	chatID := getChatID(u)
	msgs, err := ctx.GetMessages(chatID, []tg.InputMessageClass{
		&tg.InputMessageID{ID: u.CallbackQuery.MsgID},
	})
	if err != nil || len(msgs) == 0 {
		return fmt.Errorf("get message: %w", err)
	}
	msg, ok := msgs[0].(*tg.Message)
	if !ok {
		return fmt.Errorf("not a message")
	}
	media, ok := msg.Media.(*tg.MessageMediaDocument)
	if !ok {
		return fmt.Errorf("not a document")
	}
	doc, ok := media.Document.(*tg.Document)
	if !ok {
		return fmt.Errorf("empty document")
	}

	videoURL, tgVideoID, err := server.GetTgVideoPlayUrl(&server.TgVideo{Doc: doc}, avClient.LocalAddr())
	if err != nil {
		return fmt.Errorf("get video url: %w", err)
	}
	metadata := upnp.GetMetaData(msg.Message, tgVideoID, doc)
	botLogger.Info("playing tg video", zap.String("url", videoURL), zap.String("device", avClient.ServiceClient.RootDevice.Device.FriendlyName))

	if err := avClient.Stop(0); err != nil {
		botLogger.Warn("upnp stop failed", zap.Error(err))
	}
	if err := avClient.SetAVTransportURI(0, videoURL, metadata); err != nil {
		return fmt.Errorf("upnp SetAVTransportURI: %w", err)
	}
	if err := avClient.Play(0, "1"); err != nil {
		return fmt.Errorf("upnp Play: %w", err)
	}
	return nil
}

func handleTextLink(ctx *Context, u *Update) error {
	matches := linkRe.FindStringSubmatch(u.Message.Message)
	if len(matches) != 3 {
		return nil // no valid link found
	}
	peer := matches[1]
	rawMsgID := matches[2]
	msgID, err := strconv.Atoi(rawMsgID)
	if err != nil {
		return nil
	}

	botLogger.Info("resolving t.me link", zap.String("peer", peer), zap.Int("msgID", msgID))

	chatID := getChatID(u)
	localizer := lang.GetLocalizer(userLang(u))

	// Resolve the message from the link
	msg, err := resolveMessage(ctx, peer, msgID)
	if err != nil {
		botLogger.Warn("resolve link failed", zap.String("peer", peer), zap.Int("msgID", msgID), zap.Error(err))
		errMsg := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: "TgBotMsgLinkFailed", Other: "Failed to resolve link. Make sure the message is public."},
		})
		if strings.Contains(err.Error(), "CHANNEL_INVALID") || strings.Contains(err.Error(), "CHANNEL_PRIVATE") {
			errMsg = localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "TgBotMsgLinkNeedJoin", Other: "Bot needs to join this channel first. Please add the bot as a member."},
			})
		}
		ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{Message: errMsg})
		return nil
	}

	// Collect media group if the message belongs to one
	var msgs []*tg.Message
	if gid, ok := msg.GetGroupedID(); ok {
		botLogger.Info("link message is in group", zap.Int64("groupedID", gid), zap.Int("msgID", msg.ID))
		msgs = resolveMessageGroup(ctx, peer, msg.ID, msg)
	} else {
		msgs = []*tg.Message{msg}
	}
	botLogger.Info("link resolved", zap.Int("totalMsgs", len(msgs)), zap.Int("msgID", msgID))

	videos := filterVideoMessages(msgs)
	if len(videos) == 0 {
		ctx.SendMessage(chatID, &tg.MessagesSendMessageRequest{
			Message: localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{ID: "TgBotMsgLinkNoVideo", Other: "This message does not contain a playable video."},
			}),
		})
		return nil
	}

	botLogger.Info("link resolved", zap.Int("total", len(msgs)), zap.Int("videos", len(videos)))

	for _, v := range videos {
		doc, _ := v.Media.(*tg.MessageMediaDocument).Document.(*tg.Document)
		inputDoc := &tg.InputDocument{}
		inputDoc.FillFrom(doc)
		_, err = ctx.SendMedia(chatID, &tg.MessagesSendMediaRequest{
			Media: &tg.InputMediaDocument{ID: inputDoc},
			ReplyMarkup: &tg.ReplyInlineMarkup{
				Rows: []tg.KeyboardButtonRow{{
					Buttons: []tg.KeyboardButtonClass{
						&tg.KeyboardButtonCallback{
							Text:  localizer.MustLocalize(&i18n.LocalizeConfig{DefaultMessage: &i18n.Message{ID: "TgBotMsgPlay", Other: "▶️ Play"}}),
							Data:  []byte(cbPlay),
						},
					}},
				},
			},
		})
		if err != nil {
			botLogger.Error("send link video error", zap.Error(err))
		}
	}
	return nil
}

func filterVideoMessages(msgs []*tg.Message) []*tg.Message {
	var out []*tg.Message
	for _, m := range msgs {
		md, ok := m.Media.(*tg.MessageMediaDocument)
		if !ok {
			continue
		}
		doc, ok := md.Document.(*tg.Document)
		if !ok {
			continue
		}
		for _, attr := range doc.Attributes {
			if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

// resolveMessageGroup fetches all messages in the same media group as the target.
func resolveMessageGroup(ctx *Context, peer string, msgID int, target *tg.Message) []*tg.Message {
	groupedID, _ := target.GetGroupedID()
	batch := 20

	var msgs []*tg.Message
	if strings.HasPrefix(peer, "c/") {
		channelID, _ := strconv.ParseInt(peer[2:], 10, 64)
		r, err := ctx.Raw.ChannelsGetMessages(ctx.ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: channelID, AccessHash: 0},
			ID:      buildMsgIDRange(msgID, batch),
		})
		botLogger.Info("group fetch for c/ channel", zap.Int("msgID", msgID), zap.Int64("groupedID", groupedID), zap.Error(err))
		if err == nil {
			if v, ok := r.(*tg.MessagesChannelMessages); ok {
				botLogger.Info("group fetch result", zap.Int("count", len(v.Messages)))
				for _, m := range v.Messages {
					if tm, ok := m.(*tg.Message); ok {
						if gid, ok2 := tm.GetGroupedID(); ok2 {
							botLogger.Info("group msg", zap.Int("id", tm.ID), zap.Int64("gid", gid))
							if gid == groupedID {
								msgs = append(msgs, tm)
							}
						}
					}
				}
			}
		}
	} else {
		r, err := ctx.Raw.ContactsResolveUsername(ctx.ctx, &tg.ContactsResolveUsernameRequest{Username: peer})
		if err != nil {
			botLogger.Warn("resolve username for group failed", zap.Error(err))
			return []*tg.Message{target}
		}
		if ch, ok := r.Peer.(*tg.PeerChannel); ok {
			h, err := ctx.Raw.ChannelsGetMessages(ctx.ctx, &tg.ChannelsGetMessagesRequest{
				Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: 0},
				ID:      buildMsgIDRange(msgID, batch),
			})
			botLogger.Info("group fetch for public channel", zap.Int("msgID", msgID), zap.Int64("groupedID", groupedID), zap.Error(err))
			if err == nil {
				if v, ok := h.(*tg.MessagesChannelMessages); ok {
					botLogger.Info("group fetch result", zap.Int("count", len(v.Messages)))
					for _, m := range v.Messages {
						if tm, ok := m.(*tg.Message); ok {
							if gid, ok2 := tm.GetGroupedID(); ok2 {
								botLogger.Info("group msg", zap.Int("id", tm.ID), zap.Int64("gid", gid))
								if gid == groupedID {
									msgs = append(msgs, tm)
								}
							}
						}
					}
				}
			}
		}
	}

	botLogger.Info("group resolution result", zap.Int("found", len(msgs)), zap.Int("targetID", msgID), zap.Int64("groupedID", groupedID))
	if len(msgs) == 0 {
		return []*tg.Message{target}
	}
	return msgs
}

func buildMsgIDRange(center, count int) []tg.InputMessageClass {
	ids := make([]tg.InputMessageClass, 0, count)
	start := center - count/2
	for i := 0; i < count; i++ {
		id := start + i
		if id > 0 {
			ids = append(ids, &tg.InputMessageID{ID: id})
		}
	}
	return ids
}

// resolveMessage fetches a message by link: username/msgID or c/channelID/msgID.
func resolveMessage(ctx *Context, peer string, msgID int) (*tg.Message, error) {
	if strings.HasPrefix(peer, "c/") {
		// Private channel: format c/channelID/msgID
		channelID, err := strconv.ParseInt(peer[2:], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid channel ID: %w", err)
		}
		r, err := ctx.Raw.ChannelsGetMessages(ctx.ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: channelID, AccessHash: 0},
			ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: msgID}},
		})
		if err != nil {
			return nil, fmt.Errorf("channels.getMessages: %w", err)
		}
		if v, ok := r.(*tg.MessagesChannelMessages); ok {
			if msg, ok2 := v.MapMessages().First(); ok2 {
				if m, ok3 := msg.(*tg.Message); ok3 {
					return m, nil
				}
			}
		}
		return nil, fmt.Errorf("message not found")
	}

	// Public username
	r, err := ctx.Raw.ContactsResolveUsername(ctx.ctx, &tg.ContactsResolveUsernameRequest{Username: peer})
	if err != nil {
		return nil, fmt.Errorf("resolve username: %w", err)
	}
	// Use channel API for channels/supergroups
	if ch, ok := r.Peer.(*tg.PeerChannel); ok {
		c, err := ctx.Raw.ChannelsGetMessages(ctx.ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: ch.ChannelID, AccessHash: 0},
			ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: msgID}},
		})
		if err != nil {
			return nil, fmt.Errorf("channels.getMessages: %w", err)
		}
		if v, ok := c.(*tg.MessagesChannelMessages); ok {
			if msg, ok2 := v.MapMessages().First(); ok2 {
				if m, ok3 := msg.(*tg.Message); ok3 {
					return m, nil
				}
			}
		}
		return nil, fmt.Errorf("message not found")
	}
	// For users and basic groups
	msgs, err := ctx.Raw.MessagesGetMessages(ctx.ctx, []tg.InputMessageClass{
		&tg.InputMessageID{ID: msgID},
	})
	if err != nil {
		return nil, fmt.Errorf("messages.getMessages: %w", err)
	}
	switch v := msgs.(type) {
	case *tg.MessagesMessages:
		if msg, ok := v.MapMessages().First(); ok {
			if m, ok2 := msg.(*tg.Message); ok2 {
				return m, nil
			}
		}
	case *tg.MessagesChannelMessages:
		if msg, ok := v.MapMessages().First(); ok {
			if m, ok2 := msg.(*tg.Message); ok2 {
				return m, nil
			}
		}
	}
	return nil, fmt.Errorf("message not found")
}

func convInputMedia(media tg.MessageMediaClass) tg.InputMediaClass {
	md := media.(*tg.MessageMediaDocument)
	doc := md.Document.(*tg.Document)
	inputDoc := &tg.InputDocument{}
	inputDoc.FillFrom(doc)
	result := &tg.InputMediaDocument{
		Spoiler:    md.Spoiler,
		ID:         inputDoc,
		TTLSeconds: md.TTLSeconds,
	}
	result.SetFlags()
	return result
}
