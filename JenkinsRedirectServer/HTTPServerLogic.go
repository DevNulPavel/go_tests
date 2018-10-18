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

type HttpReceivedFileInfo struct {
	inputFileName string
	inputFileExt  string
	filePath      string
	fileUUID      string
}

type HttpSendFileInfo struct {
	filePath   string
	uploadName string
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func loadHtmlTemplates(contentFolder string) {
	var err error = nil
	// MainPage
	templatePath := path.Join(contentFolder, "templates/rootTemplate.html")
	mainPageTemplate, err = template.ParseFiles(templatePath)
	if checkErr(err) {
		return
	}
}

func httpRootFunc(writer http.ResponseWriter, req *http.Request) {
	// TODO: For test only
	loadHtmlTemplates(contentFolder)

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

func startHttpServer(customPort int, contentFolderLocal string) {
	port := HTTP_SERVER_PORT
	if customPort != 0 {
		port = customPort
	}

	contentFolder = contentFolderLocal

	// Static files full path
	staticFilesPath := path.Join(contentFolderLocal, "static")
	doxygenFiles := path.Join(contentFolderLocal, "doxygen_html")

	// Грузим шаблоны
	loadHtmlTemplates(contentFolderLocal)

	// Http server
	http.HandleFunc("/", httpRootFunc)
	http.HandleFunc("/redirect", httpRedirectFunc)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticFilesPath))))
	http.Handle("/doxygen_html/", http.StripPrefix("/doxygen_html/", http.FileServer(http.Dir(doxygenFiles))))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
