package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/tgbot-upnp/tgbot-upnp/bot"
	"github.com/tgbot-upnp/tgbot-upnp/config"
	"github.com/tgbot-upnp/tgbot-upnp/lang"
	"github.com/tgbot-upnp/tgbot-upnp/server"
	"github.com/tgbot-upnp/tgbot-upnp/setup"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	logger, _ := cfg.Build()
	return logger
}

// App holds the runtime state of a tgbot-upnp instance.
type App struct {
	Logger   *zap.Logger
	Context  context.Context
	HTTPPort int
	cancel   context.CancelFunc
}

// New initializes the application: loads config, runs setup wizard if needed,
// starts the bot and HTTP server. Returns a ready-to-use App.
func New() *App {
	logger := newLogger()
	lang.GetI18nBundle(logger)

	// Use TGBOT_UPNP_DATA_DIR for config/session storage
	dataDir := os.Getenv("TGBOT_UPNP_DATA_DIR")
	if dataDir == "" {
		dataDir = defaultDataDir()
	}
	_ = os.MkdirAll(dataDir, 0o700)
	_ = os.Chdir(dataDir)

	httpPort := 8080

	if config.NeedsSetup() {
		fmt.Printf("No configuration found. Opening setup wizard on http://127.0.0.1:%d\n", httpPort)
		if err := setup.Run(httpPort, lang.GetSetupStrings()); err != nil {
			fmt.Fprintln(os.Stderr, "Setup failed:", err)
			os.Exit(1)
		}
	}

	if err := config.GetConfig(logger); err != nil {
		logger.Error("Configuration error, opening setup wizard", zap.Error(err))
		fmt.Printf("Configuration error: %v\nOpening setup wizard on http://127.0.0.1:%d\n", err, httpPort)
		if err := setup.Run(httpPort, lang.GetSetupStrings()); err != nil {
			fmt.Fprintln(os.Stderr, "Setup failed:", err)
			os.Exit(1)
		}
		// Reload config after setup
		if err := config.GetConfig(logger); err != nil {
			logger.Fatal("Configuration still incomplete after setup", zap.Error(err))
		}
	}
	if p := viper.GetInt(config.HttpPort); p != 0 {
		httpPort = p
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go bot.Run(ctx,
		viper.GetInt(config.AppID),
		viper.GetString(config.ApiHash),
		viper.GetString(config.BotToken),
		viper.GetString(config.UserSession),
		".",
		viper.GetIntSlice(config.AdminIDs),
		viper.GetBool(config.AutoAdmin),
		logger,
	)

	server.Server(httpPort, logger, lang.GetSetupStrings())

	return &App{
		Logger:   logger,
		Context:  ctx,
		HTTPPort: httpPort,
		cancel:   cancel,
	}
}

// Cancel triggers graceful shutdown (useful for systray Quit).
func (a *App) Cancel() {
	a.cancel()
}

// Wait blocks until shutdown signal, then cleans up.
func (a *App) Wait() {
	<-a.Context.Done()
	server.Shutdown()
	_ = a.Logger.Sync()
}
