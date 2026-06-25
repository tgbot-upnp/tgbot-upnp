package bot

import (
	"errors"
	"strings"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// ErrStopDispatch stops further handler processing.
var ErrStopDispatch = errors.New("stop dispatch")

// Update holds the relevant parts of a Telegram update.
type Update struct {
	Message       *tg.Message
	CallbackQuery *tg.UpdateBotCallbackQuery
	UserID        int64
	UserLang      string
}

// HandlerFunc is the function signature for update handlers.
type HandlerFunc func(ctx *Context, update *Update) error

// MiddlewareFunc is a handler that runs before all others.
type MiddlewareFunc HandlerFunc

type handler struct {
	name   string
	match  func(u *Update) bool
	handle HandlerFunc
}

// Dispatcher routes incoming Telegram updates to registered handlers.
type Dispatcher struct {
	middlewares []MiddlewareFunc
	handlers    []handler
	logger      *zap.Logger
}

// NewDispatcher creates a new update dispatcher.
func NewDispatcher(logger *zap.Logger) *Dispatcher {
	return &Dispatcher{logger: logger}
}

// AddMiddleware registers a middleware that runs before all handlers.
func (d *Dispatcher) AddMiddleware(fn MiddlewareFunc) {
	d.middlewares = append(d.middlewares, fn)
}

// AddCommand registers a handler for /command messages.
func (d *Dispatcher) AddCommand(cmd string, fn HandlerFunc) {
	prefix := "/" + cmd
	d.handlers = append(d.handlers, handler{
		name:   "cmd:" + cmd,
		match:  func(u *Update) bool { return u.Message != nil && strings.HasPrefix(u.Message.Message, prefix) },
		handle: fn,
	})
}

// AddMessageText registers a handler for text messages containing a pattern.
func (d *Dispatcher) AddMessageText(contains string, fn HandlerFunc) {
	d.handlers = append(d.handlers, handler{
		name: "msg:text",
		match: func(u *Update) bool {
			return u.Message != nil && u.Message.Message != "" && strings.Contains(u.Message.Message, contains)
		},
		handle: fn,
	})
}

// AddMessageVideo registers a handler for video messages.
func (d *Dispatcher) AddMessageVideo(fn HandlerFunc) {
	d.handlers = append(d.handlers, handler{
		name: "msg:video",
		match: func(u *Update) bool {
			if u.Message == nil {
				return false
			}
			media, ok := u.Message.GetMedia()
			if !ok {
				return false
			}
			doc, ok := media.(*tg.MessageMediaDocument)
			if !ok {
				return false
			}
			d, ok := doc.Document.(*tg.Document)
			if !ok {
				return false
			}
			for _, attr := range d.Attributes {
				if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
					return true
				}
			}
			return false
		},
		handle: fn,
	})
}

// AddCallbackExact registers a handler for an exact-match callback data.
func (d *Dispatcher) AddCallbackExact(data string, fn HandlerFunc) {
	d.handlers = append(d.handlers, handler{
		name:   "cb:=" + data,
		match:  func(u *Update) bool { return u.CallbackQuery != nil && string(u.CallbackQuery.Data) == data },
		handle: fn,
	})
}

// AddCallbackPrefix registers a handler for prefix-matched callback data.
func (d *Dispatcher) AddCallbackPrefix(prefix string, fn HandlerFunc) {
	d.handlers = append(d.handlers, handler{
		name: "cb:^" + prefix,
		match: func(u *Update) bool {
			return u.CallbackQuery != nil && strings.HasPrefix(string(u.CallbackQuery.Data), prefix)
		},
		handle: fn,
	})
}

// Dispatch processes an update through middlewares then handlers.
func (d *Dispatcher) Dispatch(botCtx *Context, u *Update) {
	d.logger.Debug("received update",
		zap.Int64("userID", u.UserID),
		zap.String("userLang", u.UserLang),
		zap.Bool("hasMsg", u.Message != nil),
		zap.Bool("hasCb", u.CallbackQuery != nil),
	)
	for _, mw := range d.middlewares {
		if err := mw(botCtx, u); err != nil {
			if errors.Is(err, ErrStopDispatch) {
				return
			}
			d.logger.Error("middleware error", zap.Error(err))
			return
		}
	}
	for _, h := range d.handlers {
		if h.match(u) {
			d.logger.Debug("handler matched", zap.String("handler", h.name))
			if err := h.handle(botCtx, u); err != nil {
				d.logger.Error("handler error", zap.String("handler", h.name), zap.Error(err))
			}
			return
		}
	}
}

func extractUpdates(upd tg.UpdatesClass) []Update {
	var results []Update
	switch v := upd.(type) {
	case *tg.Updates:
		users := make(map[int64]*tg.User)
		for _, uc := range v.Users {
			if user, ok := uc.(*tg.User); ok {
				users[user.ID] = user
			}
		}
		for _, uc := range v.Updates {
			switch t := uc.(type) {
			case *tg.UpdateNewMessage, *tg.UpdateNewChannelMessage:
				var m *tg.Message
				switch vv := t.(type) {
				case *tg.UpdateNewMessage:
					m, _ = vv.Message.(*tg.Message)
				case *tg.UpdateNewChannelMessage:
					m, _ = vv.Message.(*tg.Message)
				}
				if m != nil {
					u := Update{Message: m, UserID: getUserID(m)}
					if user, ok := users[u.UserID]; ok {
						if lang, ok := user.GetLangCode(); ok {
							u.UserLang = lang
						}
					}
					results = append(results, u)
				}
			case *tg.UpdateBotCallbackQuery:
				u := Update{CallbackQuery: t, UserID: t.UserID}
				if user, ok := users[t.UserID]; ok {
					if lang, ok := user.GetLangCode(); ok {
						u.UserLang = lang
					}
				}
				results = append(results, u)
			}
		}
	case *tg.UpdateShortMessage:
		results = append(results, Update{
			Message: &tg.Message{
				ID: v.ID, Message: v.Message,
				PeerID: &tg.PeerUser{UserID: v.UserID},
				Date:   v.Date, Out: v.Out,
			},
			UserID: v.UserID,
		})
	case *tg.UpdateShortChatMessage:
		results = append(results, Update{
			Message: &tg.Message{
				ID: v.ID, Message: v.Message,
				PeerID: &tg.PeerChat{ChatID: v.ChatID},
				Date:   v.Date, Out: v.Out,
			},
			UserID: v.FromID,
		})
	}
	return results
}

func getUserID(m *tg.Message) int64 {
	if m.FromID != nil {
		if peer, ok := m.FromID.(*tg.PeerUser); ok {
			return peer.UserID
		}
	}
	// Fallback: for bot private chats, PeerID IS the user ID
	if m.PeerID != nil {
		if peer, ok := m.PeerID.(*tg.PeerUser); ok {
			return peer.UserID
		}
	}
	return 0
}
