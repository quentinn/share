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
	"path"
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

	http.Handle("/share/{id}", logreq(viewUnlockShare))

	http.Handle("/file", logreq(viewCreateFile))
	http.Handle("/file/shared", logreq(uploadFile))
	http.Handle("/file/{id}", logreq(viewRevealFile))

	http.Handle("/secret", logreq(viewCreateSecret))
	http.Handle("/secret/shared", logreq(uploadSecret))
	// http.Handle("/secret/reveal", logreq(viewRevealSecret))
	http.Handle("/secret/{id}", logreq(viewRevealSecret))
	// http.HandleFunc("/secret/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	id := r.PathValue("id")
	// 	renderTemplate(w, "view.reveal.secret.html", struct {
	// 		Name string
	// 	}{
	// 		Name: id,
	// 	})
	// })

	

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




func viewRevealSecret(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	text := readSecret(id)

	renderTemplate(w, "view.reveal.secret.html", struct {
		Name string
	}{
		Name: text,
	})
}




func viewRevealFile(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	path := readFile(id)

	renderTemplate(w, "view.reveal.secret.html", struct {
		Name string
	}{
		Name: path,
	})
}




func viewUnlockShare(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	password := readShare(id)

	renderTemplate(w, "view.unlock.share.html", struct {
		Name string
	}{
		Name: password,
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

	
	id := uuid.NewString()
	url := r.Header.Get("Referer")
	link := strings.Join([]string{url, "/", id}, "")


	// Create database entries
	createFile(id, path)



	// Display the confirmation
	renderTemplate(w, "view.confirm.file.html", struct {
		Name string
	}{
		Name: link,
	})
}




func uploadSecret(w http.ResponseWriter, r *http.Request) {


	r.ParseForm()


	id := uuid.NewString()
	shared_id := uuid.NewString()
	uri := r.Header.Get("Referer")
	url := path.Dir(uri)



	link := strings.Join([]string{"/share/", shared_id}, "")

	fmt.Println("blablabla %s", link)
	fmt.Println("blablabla %s", url)

	// Create database entries
	createSecret(id, shared_id, r.PostFormValue("mySecret"))
	



	// Display the confirmation
	renderTemplate(w, "view.confirm.secret.html", struct {
		Link string // To permit the user to click on it 
		Url string	// To permit the user to copy it
	}{
		Link: link,
		Url: url,
	})

}
