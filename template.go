package main

import (
	"html/template"
	"net/http"
	"time"
)

// Template func that checks if the current user is an admin
func IsAdmin(req *http.Request) bool {
	return WhoAmI(req).Admin
}

// UnixTime is a func that takes a timestamp and converts it
// to a unix timestamp
func UnixTime(time *time.Time) int64 {
	return time.Unix()
}

// Add func to templates
func AddTemplateFunctions(req *http.Request) template.FuncMap {
	return template.FuncMap{
		"IsAdmin":  func() bool { return IsAdmin(req) },
		"UnixTime": func(time *time.Time) int64 { return UnixTime(time) },
	}
}

// Refresh Templates recompiles the templates. We use this a lot,
// so it's better to have it in once place than in 20 places.
func RefreshTemplates(req *http.Request) *template.Template {
	return template.Must(template.New("").Funcs(AddTemplateFunctions(req)).ParseGlob("app/views/*"))
}
