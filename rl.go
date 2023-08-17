package main

import (
	"fmt"
	"strings"
	"os"
	"net/http"
	"html/template"
)

type HTMLTemplate func(any, http.ResponseWriter)

func GenerateTemplate(f string) (HTMLTemplate) {
	temp, err := template.ParseFiles(f)
	if err != nil {fmt.Println(err)}
	return func(s any, r http.ResponseWriter) { 
		err := temp.Execute(r, s)
		if err != nil {fmt.Println(err)}
	}
}

type StaticResource struct {
	Content string
	MimeType string
}

func getMimeType(f string) string {
	if strings.HasSuffix(f, ".css") {return "text/css"}
	if strings.HasSuffix(f, ".js") {return "text/javascript"}
	return "text/html"
}

func LoadStaticFile(f string) StaticResource {
	content, err := os.ReadFile(f)
	if err != nil {fmt.Println(err)}
	resource := StaticResource{Content: string(content[:]), MimeType: getMimeType(f)}
	return resource
}

type StaticBind func(http.ResponseWriter, *http.Request)

func getBindResource(res StaticResource) StaticBind {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", res.MimeType)
		fmt.Fprintf(w, res.Content)
	}	
}


