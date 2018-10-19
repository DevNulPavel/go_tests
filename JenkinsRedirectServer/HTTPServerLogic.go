package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"
)

const (
	HTTP_SERVER_PORT = 8080
)

// Templates
var mainPageTemplate *template.Template = nil
var contentFolder string = ""

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func loadHtmlTemplates(contentFolder string) {
	// MainPage
	mainTemplatePath := path.Join(contentFolder, "templates/rootTemplate.html")
	mainPageTemplate, _ = template.ParseFiles(mainTemplatePath)
}

func httpRootFunc(writer http.ResponseWriter, req *http.Request) {
	// TODO: For test only
	//loadHtmlTemplates(contentFolder)

	mainPageTemplate.Execute(writer, nil)
}

func httpRedirectFunc(w http.ResponseWriter, r *http.Request) {
	valueGet, ok := r.URL.Query()["value"]
	if ok && len(valueGet[0]) > 0 {
		value := valueGet[0]
		http.Redirect(w, r, value, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func httpServerStatFunc(w http.ResponseWriter, r *http.Request) {
	GetHardwareData(w, r)
}

func startHttpServer(customPort int, contentFolderLocal string) {
	port := HTTP_SERVER_PORT
	if customPort != 0 {
		port = customPort
	}

	contentFolder = contentFolderLocal

	// Static files full path
	staticFilesPath := path.Join(contentFolderLocal, "static")

	// Грузим шаблоны
	loadHtmlTemplates(contentFolderLocal)

	// Http server
	http.HandleFunc("/", httpRootFunc)
	http.HandleFunc("/redirect", httpRedirectFunc)
	http.HandleFunc("/server_stat", httpServerStatFunc)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticFilesPath))))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
