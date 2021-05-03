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

func startHttpServer(customPort int, contentFolderLocal string) {
	port := HTTP_SERVER_PORT
	if customPort != 0 {
		port = customPort
	}

	doxygenFiles := path.Join(contentFolderLocal, "doxygen_html")

	// Http server
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(doxygenFiles))))
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
