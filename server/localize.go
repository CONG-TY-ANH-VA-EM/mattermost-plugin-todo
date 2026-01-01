package main

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/mattermost/go-i18n/i18n"
)

//go:embed assets/i18n/*.json
var i18nFolder embed.FS

var T i18n.TranslateFunc
var locales map[string]i18n.TranslateFunc

func (p *Plugin) loadI18n() error {
	locales = make(map[string]i18n.TranslateFunc)
	
	err := fs.WalkDir(i18nFolder, "assets/i18n", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		
		data, err := i18nFolder.ReadFile(path)
		if err != nil {
			return err
		}
		
		if err := i18n.ParseTranslationFileBytes(path, data); err != nil {
			return err
		}
		
		locale := strings.TrimSuffix(filepath.Base(path), ".json")
		locales[locale] = i18n.MustTfunc(locale)
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// Default to Vietnamese per user request
	if vi, ok := locales["vi"]; ok {
		T = vi
	}
	
	return nil
}

func (p *Plugin) GetLocalizer(userID string) i18n.TranslateFunc {
	user, err := p.API.GetUser(userID)
	if err != nil {
		return T
	}
	
	locale := user.Locale
	if locale == "" {
		locale = "vi"
	}
	
	if f, ok := locales[locale]; ok {
		return f
	}
	
	// Try short locale (e.g. "en-US" -> "en")
	if strings.Contains(locale, "-") {
		short := strings.Split(locale, "-")[0]
		if f, ok := locales[short]; ok {
			return f
		}
	}
	
	return T
}

func (p *Plugin) Localize(userID string, translationID string, args map[string]interface{}) string {
	TUser := p.GetLocalizer(userID)
	if TUser == nil {
		return translationID
	}
	return TUser(translationID, args)
}
