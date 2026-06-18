package lang

import (
	"github.com/Xuanwo/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

var LocaleSystemTag language.Tag
var DefaultTag = language.English

var i18nBundle *i18n.Bundle
var logger *zap.Logger

func GetI18nBundle(globalLogger *zap.Logger) {
	logger = globalLogger

	i18nBundle = i18n.NewBundle(language.English)
	i18nBundle.AddMessages(language.English, messagesEn...)
	i18nBundle.AddMessages(language.Make("zh-Hans"), messagesZhHans...)

	localeTag, err := locale.Detect()
	if err != nil {
		logger.Error("get locale error", zap.String("err", err.Error()))
		localeTag = language.English
	} else {
		logger.Info("get locale", zap.Any("locale", localeTag))
	}
	LocaleSystemTag = getMatchTag(localeTag)
}

func getMatchTag(tag language.Tag) language.Tag {
	matcher := language.NewMatcher(i18nBundle.LanguageTags())
	matchTag, _, _ := matcher.Match(tag)
	return matchTag
}

func GetLocalizer(tag language.Tag) *i18n.Localizer {
	return i18n.NewLocalizer(i18nBundle, tag.String())
}

func GetAllSupportedTag() []language.Tag {
	return i18nBundle.LanguageTags()
}
