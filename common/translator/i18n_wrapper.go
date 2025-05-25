package translator

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/atomic"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultMsgId  = "999999"
	defaultFormat = "toml"
	unKnownMsg    = "unknown error"
)

var (
	_defaultLang       = language.English
	// 用于存储一个指向 `Translator` 类型的原子指针
	// `atomic.Pointer` 是 Go 标准库 `sync/atomic` 包中的一种类型，
	// 用于安全地操作指针，支持并发环境下的原子操作
	_defaultTranslator = atomic.Pointer[Translator]{}
)

func SetDefaultTranslator(translator *Translator) {
	// Store会存储新的指针，原来指针不使用会被go回收
	_defaultTranslator.Store(translator)
}

// *i18n.Bundle 是将 i18n.Bundle 结构体的指针嵌入到 Translator 结构体中，
// 允许 Translator 继承 i18n.Bundle 的字段和方法，并通过指针共享和修改其状态
type Translator struct {
	*i18n.Bundle
	defaultMsgs map[string]string
}

func NewTranslator(langFilePath string) (*Translator, error) {
	bundle := i18n.NewBundle(_defaultLang)
	bundle.RegisterUnmarshalFunc(defaultFormat, toml.Unmarshal)
	var defaultMsgs = map[string]string{}

	err := filepath.WalkDir(langFilePath, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if ext := strings.TrimLeft(filepath.Ext(d.Name()), "."); ext != defaultFormat {
			return nil
		}
		msgFile, err := bundle.LoadMessageFile(path)
		if err != nil {
			return err
		}

		for _, v := range msgFile.Messages {
			if v.ID == defaultMsgId {
				lang := msgFile.Tag.String()
				defaultMsgs[lang] = v.Other
				break
			}

		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &Translator{Bundle: bundle, defaultMsgs: defaultMsgs}, nil

}

func (t *Translator) Translate(lang, msgId string) string {
	localizer := i18n.NewLocalizer(t.Bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    msgId,
			Other: t.defaultMsgs[lang],
		},
	})
	var e *i18n.MessageNotFoundErr
	if err == nil || errors.As(err, &e) {
		return msg
	}
	return unKnownMsg
}

func Translate(lang, msgId string) string {
	return _defaultTranslator.Load().Translate(lang, msgId)
}
