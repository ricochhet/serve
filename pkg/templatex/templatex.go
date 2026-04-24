package templatex

import (
	"html/template"
	"net/http"

	"github.com/ricochhet/serve/pkg/embedx"
	"github.com/ricochhet/serve/pkg/syncx"
)

type Templates struct {
	*syncx.Safe[map[string]*Template]
}

type Template struct {
	*template.Template

	FS      *embedx.EmbeddedFileSystem
	Name    string
	Path    string
	FuncMap *template.FuncMap
}

func New() *Templates {
	return &Templates{
		&syncx.Safe[map[string]*Template]{},
	}
}

func (t *Templates) Register(tmpl *Template) {
	tmpl.Template = template.Must(
		template.New(tmpl.Name).Funcs(*tmpl.FuncMap).Parse(tmpl.FS.String(tmpl.Path)),
	)

	t.SetTemplate(tmpl.Name, tmpl)
}

func Render[T any](t *Template, w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = t.Execute(w, data)
}

func (t *Templates) GetTemplate(name string) (*Template, bool) {
	return syncx.GetMap(t.Safe, name)
}

func (t *Templates) SetTemplate(name string, tmpl *Template) {
	syncx.SetMap(t.Safe, name, tmpl)
}
