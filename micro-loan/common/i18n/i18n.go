package i18n

const (
	LangEnUS string = "en-US"
	LangZhCN string = "zh-CN"
	LangIdID string = "id-ID"
)

type LangItem map[string]string
type LangConfig map[string]LangItem

var langSupportConf = map[string]string{
	LangEnUS: "English",
	LangZhCN: "简体中文",
	LangIdID: "Bahasa indonesia",
}

// 翻译函数,如果没有配置语言包,显示原始数据
func T(lang, origin string) (t string) {
	t = origin

	if item, ok := langMap[origin]; ok {
		if find, has := item[lang]; has {
			t = find
		}
	}

	return
}

// 取系统支持的语言包
func LangSupportConf() map[string]string {
	return langSupportConf
}

func IsExist(lang string) bool {
	if _, ok := langSupportConf[lang]; ok {
		return true
	}

	return false
}
