package engine

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	posArgRegex = regexp.MustCompile(`%(\d+)\$[sSdD]`)
	simpleRegex = regexp.MustCompile(`%[sSdD]`)
)

func formatAndroidString(template string, args ...string) string {
	if template == "" {
		return ""
	}
	// Replace %1$s, %2$s, etc.
	res := posArgRegex.ReplaceAllStringFunc(template, func(m string) string {
		match := posArgRegex.FindStringSubmatch(m)
		if len(match) > 1 {
			idx, err := strconv.Atoi(match[1])
			if err == nil && idx >= 1 && idx <= len(args) {
				return args[idx-1]
			}
		}
		return m
	})

	// Replace simple %s or %d sequentially
	argIdx := 0
	res = simpleRegex.ReplaceAllStringFunc(res, func(m string) string {
		if argIdx < len(args) {
			val := args[argIdx]
			argIdx++
			return val
		}
		return m
	})

	return res
}

type XMLResources struct {
	XMLName xml.Name    `xml:"resources"`
	Strings []XMLString `xml:"string"`
}

type XMLString struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type LocalizationManager struct {
	Translations map[string]map[string]string
}

func NewLocalizationManager(dataDir string) *LocalizationManager {
	lm := &LocalizationManager{
		Translations: make(map[string]map[string]string),
	}
	lm.loadAll(dataDir)
	return lm
}

func (lm *LocalizationManager) loadAll(dataDir string) {
	locales := map[string]string{
		"pt":    "values",
		"en":    "values-en",
		"es":    "values-es",
		"fr":    "values-fr",
		"de":    "values-de",
		"la":    "values-la",
		"pt-br": "values-pt-rBR",
	}

	for lang, folder := range locales {
		path := filepath.Join(dataDir, folder, "strings.xml")
		lm.Translations[lang] = lm.parseStrings(path)
	}
}

func (lm *LocalizationManager) parseStrings(path string) map[string]string {
	stringsMap := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		return stringsMap
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return stringsMap
	}

	var resources XMLResources
	err = xml.Unmarshal(byteValue, &resources)
	if err != nil {
		fmt.Printf("Error unmarshaling %s: %v\n", path, err)
		return stringsMap
	}

	for _, s := range resources.Strings {
		val := s.Value
		val = strings.ReplaceAll(val, `\'`, `'`)
		val = strings.ReplaceAll(val, `\"`, `"`)
		stringsMap[s.Name] = val
	}

	return stringsMap
}

func (lm *LocalizationManager) GetTranslations(lang string) map[string]string {
	langNormalized := strings.ReplaceAll(strings.ToLower(lang), "_", "-")
	if trans, ok := lm.Translations[langNormalized]; ok && len(trans) > 0 {
		return trans
	}

	base := strings.Split(langNormalized, "-")[0]
	if trans, ok := lm.Translations[base]; ok && len(trans) > 0 {
		return trans
	}

	if strings.HasPrefix(langNormalized, "pt") {
		if trans, ok := lm.Translations["pt-br"]; ok && len(trans) > 0 {
			return trans
		}
		if trans, ok := lm.Translations["pt"]; ok && len(trans) > 0 {
			return trans
		}
	}

	if trans, ok := lm.Translations["en"]; ok && len(trans) > 0 {
		return trans
	}
	if trans, ok := lm.Translations["pt"]; ok && len(trans) > 0 {
		return trans
	}

	for _, trans := range lm.Translations {
		if len(trans) > 0 {
			return trans
		}
	}

	return make(map[string]string)
}
