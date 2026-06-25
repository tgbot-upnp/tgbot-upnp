package bot

import (
	"context"
	"encoding/base64"

	"github.com/gotd/td/clock"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
	"github.com/tgbot-upnp/tgbot-upnp/server"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

// userAPI provides optional user-scoped API for accessing restricted channels.
var userAPI *tg.Client

// createUserClient initializes a Telegram client from a user session string.
// The client runs in a goroutine and sets userAPI on success.
func createUserClient(ctx context.Context, appID int, apiHash, userSession string, autoAdmin bool) {
	if userSession == "" {
		return
	}
	data, err := base64.StdEncoding.DecodeString(userSession)
	if err != nil {
		botLogger.Warn("decode user session failed", zap.Error(err))
		return
	}
	var userSess session.StorageMemory
	_ = userSess.StoreSession(ctx, data)
	userClient := telegram.NewClient(appID, apiHash, telegram.Options{
		Resolver:       dcs.Plain(dcs.PlainOptions{Dial: proxy.Direct.DialContext}),
		SessionStorage: &userSess,
		Clock:          clock.System,
	})
	go func() {
		if err := userClient.Run(ctx, func(ctx context.Context) error {
			userAPI = userClient.API()
			server.SetUserAPI(userAPI)
			if autoAdmin {
				if self, err := userClient.Self(ctx); err == nil {
					autoAdminUserID.Store(self.ID)
					botLogger.Info("auto-admin user set", zap.Int64("userID", self.ID))
				} else {
					botLogger.Warn("auto-admin: failed to get self", zap.Error(err))
				}
			}
			botLogger.Info("user session loaded")
			<-ctx.Done()
			return nil
		}); err != nil {
			botLogger.Warn("user session client failed", zap.Error(err))
		}
	}()
}
