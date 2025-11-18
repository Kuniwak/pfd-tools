package locale

import "fmt"

type Locale string

const (
	LocaleJa Locale = "ja"
	LocaleEn Locale = "en"
)

func (l Locale) String() string {
	return string(l)
}

func Parse(s string) (Locale, error) {
	switch s {
	case LocaleJa.String():
		return LocaleJa, nil
	case LocaleEn.String(), "":
		return LocaleEn, nil
	}
	return LocaleEn, fmt.Errorf("locale.ParseLocale: unknown locale: %q", s)
}
