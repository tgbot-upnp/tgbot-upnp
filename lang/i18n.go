package lang

import (
	"embed"
	"github.com/BurntSushi/toml"
	"github.com/Xuanwo/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

var LocaleSystemTag language.Tag
var DefaultTag = language.English

//go:embed *.toml
var LocaleFS embed.FS

var i18nBundle *i18n.Bundle
var logger *zap.Logger

func GetI18nBundle(globalLogger *zap.Logger) {
	logger = globalLogger
	getSupportedLanguage()
	localeTag, err := locale.Detect()
	if err != nil {
		logger.Error("get locale error", zap.String("err", err.Error()))
		localeTag = language.English
	} else {
		logger.Info("get locale", zap.Any("locale", localeTag))
	}
	LocaleSystemTag = getMatchTag(localeTag)
}
func getSupportedLanguage() {
	i18nBundle = i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	dirEntries, _ := LocaleFS.ReadDir(".")
	for _, dirEntry := range dirEntries {
		_, err := i18nBundle.LoadMessageFileFS(LocaleFS, dirEntry.Name())
		if err != nil {
			logger.Fatal("load message File success", zap.String("file", dirEntry.Name()))
		}
		logger.Info("load message File success", zap.String("file", dirEntry.Name()))
	}
}
func getMatchTag(tag language.Tag) (matchTag language.Tag) {
	matcher := language.NewMatcher(i18nBundle.LanguageTags())
	matchTag, _, _ = matcher.Match(tag)
	return
}
func GetLocalizer(tag language.Tag) (localizer *i18n.Localizer) {
	localizer = i18n.NewLocalizer(i18nBundle, tag.String())
	return
}
func GetAllSupportedTag() []language.Tag {
	return i18nBundle.LanguageTags()
}
