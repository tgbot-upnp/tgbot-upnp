package bot

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/gotd/td/tg"
)

var (
	sharedRand   = rand.New(rand.NewSource(time.Now().UnixNano()))
	sharedRandMu sync.Mutex
)

// Context wraps the raw Telegram API client with convenience helpers.
type Context struct {
	Raw  *tg.Client
	Self *tg.User
	ctx  context.Context
}

func newBotContext(ctx context.Context, raw *tg.Client, self *tg.User) *Context {
	return &Context{ctx: ctx, Raw: raw, Self: self}
}

func (c *Context) randomID() int64 {
	sharedRandMu.Lock()
	defer sharedRandMu.Unlock()
	return sharedRand.Int63()
}

// SendMessage sends a message to the given chat.
func (c *Context) SendMessage(chatID int64, req *tg.MessagesSendMessageRequest) (*tg.Message, error) {
	if req == nil {
		req = &tg.MessagesSendMessageRequest{}
	}
	req.RandomID = c.randomID()
	req.Peer = &tg.InputPeerUser{UserID: chatID}
	u, err := c.Raw.MessagesSendMessage(c.ctx, req)
	if err != nil {
		return nil, err
	}
	return messageFromUpdates(u), nil
}

// SendMedia sends a media message to the given chat.
func (c *Context) SendMedia(chatID int64, req *tg.MessagesSendMediaRequest) (*tg.Message, error) {
	if req == nil {
		req = &tg.MessagesSendMediaRequest{}
	}
	req.RandomID = c.randomID()
	req.Peer = &tg.InputPeerUser{UserID: chatID}
	u, err := c.Raw.MessagesSendMedia(c.ctx, req)
	if err != nil {
		return nil, err
	}
	return messageFromUpdates(u), nil
}

// EditMessage edits a message in the given chat.
func (c *Context) EditMessage(chatID int64, req *tg.MessagesEditMessageRequest) error {
	if req == nil {
		req = &tg.MessagesEditMessageRequest{}
	}
	req.Peer = &tg.InputPeerUser{UserID: chatID}
	_, err := c.Raw.MessagesEditMessage(c.ctx, req)
	return err
}

// AnswerCallback answers a callback query.
func (c *Context) AnswerCallback(req *tg.MessagesSetBotCallbackAnswerRequest) (bool, error) {
	if req == nil {
		req = &tg.MessagesSetBotCallbackAnswerRequest{}
	}
	return c.Raw.MessagesSetBotCallbackAnswer(c.ctx, req)
}

// GetMessages fetches messages by ID from a chat.
func (c *Context) GetMessages(chatID int64, msgIDs []tg.InputMessageClass) ([]tg.MessageClass, error) {
	messages, err := c.Raw.MessagesGetMessages(c.ctx, msgIDs)
	if err != nil {
		return nil, err
	}
	switch v := messages.(type) {
	case *tg.MessagesMessages:
		return v.Messages, nil
	case *tg.MessagesChannelMessages:
		result := make([]tg.MessageClass, 0, len(v.Messages))
		for _, m := range v.Messages {
			result = append(result, m)
		}
		return result, nil
	case *tg.MessagesMessagesSlice:
		return v.Messages, nil
	default:
		return nil, nil
	}
}

// messageFromUpdates extracts the first message from an UpdatesClass.
func messageFromUpdates(u tg.UpdatesClass) *tg.Message {
	switch v := u.(type) {
	case *tg.Updates:
		for _, upd := range v.Updates {
			if msg, ok := upd.(*tg.UpdateNewMessage); ok {
				if m, ok := msg.Message.(*tg.Message); ok {
					return m
				}
			}
		}
	case *tg.UpdateShortMessage:
		return &tg.Message{
			ID:      v.ID,
			Message: v.Message,
			PeerID:  &tg.PeerUser{UserID: v.UserID},
			Date:    v.Date,
			Out:     v.Out,
		}
	case *tg.UpdateShortChatMessage:
		return &tg.Message{
			ID:      v.ID,
			Message: v.Message,
			PeerID:  &tg.PeerChat{ChatID: v.ChatID},
			Date:    v.Date,
			Out:     v.Out,
		}
	}
	return nil
}
