package gopicotts

//go:generate stringer -type=Language
type Language byte

const (
	LanguageEnUS Language = iota
	LanguageEnGB
	LanguageDe
	LanguageEs
	LanguageFr
	LanguageIt
)

type langInfo struct {
	iso3                 string
	countryIso3          string
	supportedLang        string
	internalLang         string
	internalTaLingware   string
	internalSgLingware   string
	internalUtppLingware string
}

var enUS = langInfo{"eng", "USA", "en-US", "en-US", "en-US_ta.bin", "en-US_lh0_sg.bin", "en-US_utpp.bin"}
var enGB = langInfo{"eng", "GBR", "en-GB", "en-GB", "en-GB_ta.bin", "en-GB_kh0_sg.bin", "en-GB_utpp.bin"}
var deu = langInfo{"deu", "DEU", "de-DE", "de-DE", "de-DE_ta.bin", "de-DE_gl0_sg.bin", "de-DE_utpp.bin"}
var spa = langInfo{"spa", "SPA", "es-ES", "es-ES", "es-ES_ta.bin", "es-ES_zl0_sg.bin", "es-ES_utpp.bin"}
var fra = langInfo{"fra", "FRA", "fr-FR", "fr-FR", "fr-FR_ta.bin", "fr-FR_nk0_sg.bin", "fr-FR_utpp.bin"}
var ita = langInfo{"ita", "ITA", "it-IT", "it-IT", "it-IT_ta.bin", "it-IT_cm0_sg.bin", "it-IT_utpp.bin"}

func SupportedLanguages() []string {
	return []string{"en-US", "en-GB", "de-DE", "es-ES", "fr-FR", "it-IT"}
}

func ParseLanguageName(name string) Language {
	switch name {
	case "en-GB":
		return LanguageEnGB
	case "de-DE":
		return LanguageDe
	case "es-ES":
		return LanguageEs
	case "fr-FR":
		return LanguageFr
	case "it-IT":
		return LanguageIt
	case "en-US":
		fallthrough
	default:
		return LanguageEnUS
	}
}
func getLangInfo(l Language) langInfo {
	switch l {
	case LanguageEnUS:
		return enUS
	case LanguageEnGB:
		return enGB
	case LanguageDe:
		return deu
	case LanguageEs:
		return spa
	case LanguageFr:
		return fra
	case LanguageIt:
		return ita
	default:
		return langInfo{}
	}
}
