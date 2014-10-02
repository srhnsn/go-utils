package i18n

import (
	"html/template"
	"net/http"

	"github.com/nicksnyder/go-i18n/i18n"
	i18nLanguage "github.com/nicksnyder/go-i18n/i18n/language"
	i18nTranslation "github.com/nicksnyder/go-i18n/i18n/translation"
	"github.com/srhnsn/go-utils/log"
	"gopkg.in/yaml.v2"
)

type DataWrappedTranslateFunc func(translationID string, args ...interface{}) string
type UnescapedTranslateFunc func(translationID string, args ...interface{}) template.HTML

type FutureTranslateFunc func(translationID string, args ...interface{}) FutureTranslation
type FutureTranslation func() template.HTML

type I18nConfig struct {
	Asset           func(name string) ([]byte, error)
	AssetDir        func(name string) ([]string, error)
	DefaultLanguage string
	Languages       []string
}

var i18nConfig I18nConfig

func InitI18n(config I18nConfig) {
	i18nConfig = config
	// names, err := config.AssetDir("locale")

	// if err != nil {
	// 	log.Error.Fatalf("Could not read locale directory: %s", err)
	// }

	names := config.Languages

	for _, name := range names {
		filename := "locale/" + name + "/messages.yaml"
		fileContent, err := config.Asset(filename)

		if err != nil {
			log.Warning.Printf("Could not read messages file \"%s\": %s", filename, err)
			continue
		}

		languages := i18nLanguage.Parse(name)

		if languages == nil {
			log.Warning.Printf("Not a parseable language: %s", name)
			continue
		}

		language := languages[0]

		data := make(map[string]interface{})
		yaml.Unmarshal(fileContent, data)

		count := 0

		for key, value := range data {
			translation, err := i18nTranslation.NewTranslation(map[string]interface{}{
				"id":          key,
				"translation": getTypedTranslationValue(value),
			})

			if err != nil {
				log.Error.Printf(`Language key "%s" in %s is invalid: %s`, key, filename, err)
				continue
			}

			i18n.AddTranslation(language, translation)
			count++
		}

		log.Trace.Printf("Loaded %d translations for \"%s\"", count, name)
	}
}

func GetTranslationFunc(r *http.Request) i18n.TranslateFunc {
	var cookieLang string
	cookie, err := r.Cookie("lang")

	if err == nil {
		cookieLang = cookie.Value
	} else {
		cookieLang = ""
	}

	acceptLang := r.Header.Get("Accept-Language")

	T, err := i18n.Tfunc(cookieLang, acceptLang, i18nConfig.DefaultLanguage)

	if err != nil {
		log.Error.Fatalf("Could not load translation function: %s", err)
	}

	return T
}

func GetDataWrappedTranslateFunc(T i18n.TranslateFunc, data map[string]interface{}) DataWrappedTranslateFunc {
	var DataWrappedT DataWrappedTranslateFunc = func(translationID string, args ...interface{}) string {
		var out string
		argsLen := len(args)

		if argsLen == 0 {
			out = T(translationID, map[string]interface{}(data))
		} else if argsLen == 1 {
			out = T(translationID, args[0], map[string]interface{}(data))
		} else {
			log.Error.Panicf(`Translation ID "%s" called with too many (%d) arguments`, translationID, argsLen)
		}

		return out
	}

	return DataWrappedT
}

func GetFutureTranslateFunc(T UnescapedTranslateFunc) FutureTranslateFunc {
	var FutureT FutureTranslateFunc = func(translationID string, args ...interface{}) FutureTranslation {
		var f FutureTranslation = func() template.HTML {
			return T(translationID, args...)
		}

		return f
	}

	return FutureT
}

func GetUnescapedTranslatFunc(T DataWrappedTranslateFunc) UnescapedTranslateFunc {
	var UnescapedT UnescapedTranslateFunc = func(translationID string, args ...interface{}) template.HTML {
		out := T(translationID, args...)
		return template.HTML(out)
	}

	return UnescapedT
}

func getTypedTranslationValue(value interface{}) interface{} {
	valueStr, ok := value.(string)

	if ok {
		return valueStr
	}

	valueIntf, ok := value.(map[interface{}]interface{})

	if !ok {
		log.Error.Fatalf("getTypedTranslationValue() got a really strange value: %s", value)
	}

	valueStrMap := make(map[string]interface{})

	for k, v := range valueIntf {
		k := k.(string)
		valueStrMap[k] = v
	}

	return valueStrMap
}
