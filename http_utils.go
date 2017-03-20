package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/silentsokolov/go-sleep/log"
)

//go:generate go-bindata -o templates.go ./templates/...

var (
	templates       = map[string]*template.Template{}
	templateFuncMap = template.FuncMap{
		"CheckExistsTime": func(i *time.Time) bool {
			if i == nil {
				return false
			}
			return true
		},
	}
)

func loadTemplates() {
	filenames, err := filepath.Glob("templates/*")
	if err != nil {
		log.Fatalln(err)
	}
	for _, filename := range filenames {
		name := filepath.Base(filename)
		asset, err := Asset("templates/" + name)
		if err != nil {
			log.Fatal(err)
		}
		templates[name] = template.Must(template.New(name).Funcs(templateFuncMap).Parse(string(asset)))
	}
}

func responseJSON(w http.ResponseWriter, status int, context interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	js, err := json.Marshal(context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(js)
}

func responseHTML(w http.ResponseWriter, status int, nameTmp string, context interface{}) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if len(templates) == 0 {
		loadTemplates()
	}

	template := templates[nameTmp]
	if err := template.Execute(w, context); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
