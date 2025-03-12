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
	// "net/url"
	"encoding/json"
    "crypto/sha256"
	// "math"

)




type App struct {
	Port string
}




func main() {

	createDatabase()

	server := App{
		Port: env("PORT", "8080"),
	}
	server.Start()
}




func (a *App) Start() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.Handle("/", logreq(viewIndex))

	http.Handle("/file", logreq(viewCreateFile))								// Form to create a share
	http.Handle("/file/shared", logreq(uploadFile))								// Confirmation + display the link of the share to the creator
	http.Handle("/file/{id}", logreq(viewRevealFile))							// Reveal the share after unlocked with password
	
	http.Handle("/secret", logreq(viewCreateSecret))							// Form to create a share
	http.Handle("/secret/shared", logreq(uploadSecret))							// Confirmation + display the link of the share to the creator
	http.Handle("/secret/{id}", logreq(viewRevealSecret))						// Reveal the share after unlocked with password

	http.Handle("/share/{id}", logreq(viewUnlockShare))							// Ask for password to unlock the share
	http.Handle("/share/unlock", logreq(unlockShare))							// Non browsable url - verify password to unlock the share



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

	// Generate a token that will permit to prevent unwanted record to database due to browse the upload URL without using the form
	// The trick is that this token is used from an hidden input on the HTML form, and if it's empty it means we're not using the form
	token := generatePassword()

	renderTemplate(w, "view.create.secret.html", struct {
		TokenAvoidRefresh string
	}{
		TokenAvoidRefresh: token,
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

	renderTemplate(w, "view.reveal.file.html", struct {
		Name string
	}{
		Name: path,
	})
}




func viewUnlockShare(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")
	password_database := getSharePassword(id)
	

	renderTemplate(w, "view.unlock.share.html", struct {
		Id string
		Password string
	}{
		Id: id,
		Password: password_database,
	})
}




func unlockShare(w http.ResponseWriter, r *http.Request)  {

		r.ParseForm()


		url := r.Header.Get("Referer")
		idToUnlock := url[len(url)-36:] // Just get the last 36 char of the url because the IDs are 36 char length


		givenPasswordHash := r.FormValue("givenPasswordHash")


		sharePassword := getSharePassword(idToUnlock)
		hash := sha256.New()
		hash.Write([]byte(sharePassword))
		sharePasswordHash := fmt.Sprintf("%x", []byte(hash.Sum(nil)))




		if givenPasswordHash == sharePasswordHash {
			data := map[string]interface{}{
				// "idToUnlock":    idToUnlock,
				"sharePasswordHash":	sharePasswordHash,
				"sharePassword":		getSharePassword(idToUnlock),		// return the password of the share to the JS formData (this permit to avoid writing it in DOM)
			}
			
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Printf("could not marshal json: %s\n", err)
				return
			}
		
			w.Write(jsonData) // write JSON to JS

		} else {
			fmt.Printf("password hash mismatch\n")
		}

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


	tokenAvoidRefresh := r.PostFormValue("TokenAvoidRefresh")
	if tokenAvoidRefresh != "" {

		id := uuid.NewString()
		shared_id := uuid.NewString()
		uri := r.Header.Get("Referer")		// Entire path 'http://domain:port/node1/node2/etc.../'
		url := path.Dir(uri)				// Only the 'http://domain:port' part
		link := strings.Join([]string{"/share/", shared_id}, "")


		// Create database entries
		createSecret(id, shared_id, r.PostFormValue("mySecret"))


		// Display the confirmation
		renderTemplate(w, "view.confirm.secret.html", struct {
			Link string				// To permit the user to click on it 
			Url string				// To permit the user to copy it
			Password string			// To permit the user to copy it
		}{
			Link: link,
			Url: url,
			Password: getSharePassword(shared_id),
		})
	}
}
