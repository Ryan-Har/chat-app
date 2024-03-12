package main

import (
	"html/template"
	"net/http"
	"os"
)

func main() {

	data := struct {
		ChatHost string
		ChatPort string
	}{
		ChatHost: os.Getenv("chatHost"),
		ChatPort: os.Getenv("chatPort"),
	}

	tmpl, err := template.ParseFiles("web/index.html")
	if err != nil {
		panic(err)
	}

	// Handle the root URL
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Execute the template, passing in the data
		err := tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Serve static files
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	http.ListenAndServe(":8080", nil)
}
