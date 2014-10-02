package webapps

import (
	"html/template"
	"net/http"

	"github.com/srhnsn/go-utils/i18n"
	"github.com/srhnsn/go-utils/log"
	"github.com/yosssi/ace"
)

const basePath = "views"
const baseTemplateName = "base"

type TemplateData map[string]interface{}

type TemplateConfig struct {
	Asset func(name string) ([]byte, error)
	Root  string
}

var templateConfig TemplateConfig

func GetTemplate(name string, r *http.Request) *template.Template {
	data := GetTemplateData(r)

	funcMap := template.FuncMap{
		"static": GetStaticPath,
		"T":      data["UnsafeT"],
	}

	tpl, err := ace.Load(baseTemplateName, name, &ace.Options{
		Asset:   templateConfig.Asset,
		BaseDir: basePath,
		FuncMap: funcMap,
	})

	// Specify again here because tpl is cached
	tpl.Funcs(funcMap)

	if err != nil {
		log.Error.Fatalf("Could not load template %s: %s", name, err)
	}

	return tpl
}

func InitTemplates(config TemplateConfig) {
	templateConfig = config
	initStaticPath(config.Root)
}

func getDefaultTemplateData(r *http.Request) TemplateData {
	data := make(TemplateData)

	data["csrf"] = getCsrfHTML(r)
	data["request_uri"] = r.RequestURI
	data["root"] = templateConfig.Root

	return data
}

func sanitizeData(data TemplateData) {
	futureTs := make(map[string]i18n.FutureTranslation)

	for key, value := range data {
		switch t := value.(type) {
		case string:
			data[key] = template.HTML(template.HTMLEscapeString(t))
		case i18n.FutureTranslation:
			futureTs[key] = t
		}
	}

	for key, value := range futureTs {
		data[key] = value()
	}
}
