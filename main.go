package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"io"
)

type App struct {
	Port string
}

func main() {
	server := App{
		Port: env("PORT", "8080"),
	}
	server.Start()
}

func (a *App) Start() {
	http.Handle("/", logreq(viewIndex))
	http.Handle("/file", logreq(viewFile))
	http.Handle("/secret", logreq(viewSecret))
	http.HandleFunc("/upload", uploadFile)


	addr := fmt.Sprintf(":%s", a.Port)
	log.Printf("Starting app on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}


func env(key, adefault string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return adefault
	}
	return val
}

func logreq(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s", r.URL.Path)

		f(w, r)
	})
}

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	// This is inefficient - it reads the templates from the filesystem every
	// time. This makes it much easier to develop though, so we can edit our
	// templates and the changes will be reflected without having to restart
	// the app.
	t, err := template.ParseGlob("templates/*.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error %s", err.Error()), 500)
		return
	}

	err = t.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error %s", err.Error()), 500)
		return
	}
}



func viewIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "view.index.html", struct {
		Name string
	}{
		Name: "name to fill",
	})
}



func viewFile(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "view.file.html", struct {
		Name string
	}{
		Name: "name to fill",
	})
}



func viewSecret(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "view.secret.html", struct {
		Name string
	}{
		Name: "name to fill",
	})
}



func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	r.ParseMultipartForm(10 << 20)

	// Get handler for filename, size and headers
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create file
	dst, err := os.Create(handler.Filename)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy the uploaded file to the created file on the filesystem
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}