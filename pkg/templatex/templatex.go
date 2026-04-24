package templatex

import (
	"html/template"
	"net/http"

	"github.com/ricochhet/serve/pkg/embedx"
	"github.com/ricochhet/serve/pkg/syncx"
)

type templates map[string]*template.Template

type Templates struct {
	*syncx.Safe[templates]
}

func New() *Templates {
	return &Templates{
		&syncx.Safe[templates]{},
	}
}

type Template struct {
	FS      *embedx.EmbeddedFileSystem
	Name    string
	Path    string
	FuncMap *template.FuncMap
}

func (t *Templates) Register(tmpl *Template) {
	t.SetTemplate(
		tmpl.Name,
		template.Must(
			template.New(tmpl.Name).Funcs(*tmpl.FuncMap).Parse(tmpl.FS.String(tmpl.Path)),
		),
	)
}

func Render[T any](t *Templates, name string, w http.ResponseWriter, status int, data T) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = t.GetTemplate(name).Execute(w, data)
}

func (t *Templates) GetTemplate(name string) *template.Template {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	m := t.Get()
	if m == nil {
		return nil
	}

	return (*m)[name]
}

func (t *Templates) SetTemplate(name string, tmpl *template.Template) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	m := t.Get()
	if m == nil {
		t.Set(&templates{name: tmpl})
		return
	}

	(*m)[name] = tmpl
}
