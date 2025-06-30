package render

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"reflect"
	"time"

	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/config"
	"github.com/mcgigglepop/brilliant-inferno-ruby/server/internal/models"
	"github.com/justinas/nosurf"
)

var app *config.AppConfig
var pathToTemplates = "./templates"

var functions = template.FuncMap{
	"humanDate":        HumanDate,
	"formatDate":       FormatDate,
	"formatStringDate": FormatStringDate,
	"add":              Add,
	"seq": func(n int) []int {
		s := make([]int, n)
		for i := 0; i < n; i++ {
			s[i] = i
		}
		return s
	},
	"len": func(x interface{}) int {
		return reflect.ValueOf(x).Len()
	},
	"lt": func(a, b int) bool {
		return a < b
	},
}

func NewRenderer(a *config.AppConfig) {
	app = a
}

func HumanDate(t time.Time) string {
	return t.Format("1/2/2006, 3:04 PM")
}

func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

func FormatStringDate(date string) string {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return "Invalid date"
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

func Add(a, b int) int {
	return a + b
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		return errors.New("can't get template from cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
		return err
	}
	return nil
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		layouts, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if len(layouts) > 0 {
			ts, err = ts.ParseFiles(layouts...)
			if err != nil {
				return myCache, err
			}
		}

		partials, err := filepath.Glob(fmt.Sprintf("%s/partials/*.partial.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}
		if len(partials) > 0 {
			ts, err = ts.ParseFiles(partials...)
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}