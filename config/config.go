package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"strings"
)

const (
	EnvPrefix = "TELEGRAM"
	AppID     = "APP_ID"
	ApiHash   = "API_HASH"
	BotToken  = "BOT_TOKEN"
	AdminID   = "ADMIN_ID"
	AdminIDs  = "ADMIN_IDS"
	HttpPort  = "HTTP_PORT"
)

func GetConfig(logger *zap.Logger) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault(HttpPort, 8080)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("Configuration file not exist,use environment variables", zap.String("err", err.Error()))
		} else {
			logger.Info("Configuration file read error,use environment variables", zap.String("err", err.Error()))
		}
	}
	viper.SetEnvPrefix(EnvPrefix)
	viper.AutomaticEnv()
	if viper.GetInt(AppID) == 0 || viper.GetString(ApiHash) == "" || viper.GetString(BotToken) == "" || viper.GetString(AdminID) == "" {
		logger.Fatal("Configuration is incomplete,please check")
	}
	viper.Set(AdminIDs, strings.Split(viper.GetString(AdminID), ","))
	if len(viper.GetIntSlice(AdminIDs)) == 0 {
		logger.Fatal("Admin ID is incomplete,please check")
	}
}
