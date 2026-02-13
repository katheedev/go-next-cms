package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Bundle struct {
	messages map[string]map[string]string
}

func Load(dir string) (*Bundle, error) {
	langs := []string{"en", "si", "ta"}
	b := &Bundle{messages: map[string]map[string]string{}}
	for _, lang := range langs {
		file := filepath.Join(dir, lang+".json")
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		m := map[string]string{}
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		b.messages[lang] = m
	}
	return b, nil
}

func (b *Bundle) T(lang, key string) string {
	if m, ok := b.messages[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := b.messages["en"]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

func Supported(lang string) bool { return lang == "en" || lang == "si" || lang == "ta" }
