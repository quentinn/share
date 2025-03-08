package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"io"
	"path/filepath"
	"errors"
	// "github.com/satori/go.uuid"
	"github.com/google/uuid"

	"strings"
)




type App struct {
	Port string
}




// type Share struct {
//     Id string
//     Password string
// }




func main() {

	createDatabase()

	server := App{
		Port: env("PORT", "8080"),
	}
	server.Start()
}




func (a *App) Start() {
	http.Handle("/", logreq(viewIndex))
	http.Handle("/file", logreq(viewCreateFile))
	http.Handle("/secret", logreq(viewCreateSecret))
	http.Handle("/share/file", logreq(uploadFile))
	http.Handle("/share/secret", logreq(uploadSecret))
	// http.HandleFunc("/share/file", uploadFile)
	// http.HandleFunc("/share/secret", uploadSecret)


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
		// log.Printf("path: %s", r.URL.Path)
		log.Printf("url: %s", r.Header.Get("Referer"))

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




func viewCreateFile(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "view.create.file.html", struct {
		Name string
	}{
		Name: "name to fill",
	})
}




func viewCreateSecret(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "view.create.secret.html", struct {
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
		fmt.Println("Error retrieving the file")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded file: %+v\n", handler.Filename)
	fmt.Printf("File size: %+v\n", handler.Size)
	fmt.Printf("MIME header: %+v\n", handler.Header)


	// Create destination directory
	dir := "uploads"
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}


	// Create file
	path := filepath.Join(dir, filepath.Base(handler.Filename))
	dst, err := os.Create(path)
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

	
	// Create database entries
	// createShare()
	// id := reflect.ValueOf(uuid.Must(uuid.NewV4()))
	id := uuid.NewString()
	createFile(id, path)



	// Display the confirmation
	renderTemplate(w, "view.confirm.file.html", struct {
		Name string
	}{
		Name: "name to fill",
	})
}




func uploadSecret(w http.ResponseWriter, r *http.Request) {


	r.ParseForm()


	id := uuid.NewString()
	url := r.Header.Get("Referer")
	link := strings.Join([]string{url, "/", id}, "")


	// Create database entries
	// createShare()
	createSecret(id, r.PostFormValue("mySecret"))
	



	// Display the confirmation
	renderTemplate(w, "view.confirm.secret.html", struct {
		Name string
	}{
		// Name: readShare(createShare()),
		Name: link,
	})
}
