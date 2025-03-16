package main

import (
	"fmt"
	"log"
	"os"
	"io"
	"errors"
	"strings"
	"path"
	"path/filepath"
	"html/template"
	"net/http"
	"encoding/json"
	"crypto/sha256"

	"github.com/google/uuid"

)




type App struct {
	Port string
}




func main() {


	server := App{
		Port: env("PORT", "8080"),
	}
	

	args := []string(os.Args[1:])
	if len(args) >= 1 {
		// go run share web
		if string(os.Args[1]) == "web" {
			go periodicClean()	// Goroutine to clean expired shares
			os.Setenv("DELETE_DB_ON_NEXT_START", "false")
			createDatabase()
			server.Start()

		// go run share reset
		} else if string(os.Args[1]) == "reset" {
			fmt.Println("Resetting database")
			os.Setenv("DELETE_DB_ON_NEXT_START", "true")
			createDatabase()

		// go run share delete <share_id>
		} else if string(os.Args[1]) == "delete" {
			if len(args) > 1 {
				shareId := string(os.Args[2])
				fmt.Println("Deleting share '%s'", shareId)
				deleteShare(shareId)
			} else {
				fmt.Println("Please provide a share id")
			}

		// go run share backup
		} else if string(os.Args[1]) == "backup" {
			backupFile("sqlite.db")

		// go run share help
		} else if string(os.Args[1]) == "help" {
			fmt.Println("Share is a web service that permit to securely share files and secrets to anyone")
			fmt.Println("")
			fmt.Println("Usage:")
			fmt.Println(" go run share web                  start web server")
			fmt.Println(" go run share reset                delete database, it will be recreated next web server start")
			fmt.Println(" go run share delete <share_id>    delete the share which belongs to the given id (also delete shared files if any)")
			fmt.Println(" go run share backup               duplicate database (!does not backups shared files!)")
			fmt.Println("")
			fmt.Println("https://github.com/ggtrd/share")

		// go run share <any wrong option>
		} else {
			fmt.Println("error: unknown command")
			fmt.Println("use 'go run share help' to display usage")
			fmt.Println("")
		}

	// go run share
	} else {
		fmt.Println("error: empty argument")
		fmt.Println("use 'go run share help' to display usage")
		fmt.Println("")
	}



}




func (a *App) Start() {

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	http.Handle("/", logReq(viewIndex))

	http.Handle("/file", logReq(viewCreateFile))								// Form to create a share
	http.Handle("/file/shared", logReq(uploadFile))								// Confirmation + display the link of the share to the creator
	
	http.Handle("/secret", logReq(viewCreateSecret))							// Form to create a share
	http.Handle("/secret/shared", logReq(uploadSecret))							// Confirmation + display the link of the share to the creator

	http.Handle("/share/{id}", logReq(viewUnlockShare))							// Ask for password to unlock the share
	http.Handle("/share/unlock", logReq(unlockShare))							// Non browsable url - verify password to unlock the share
	http.Handle("/share/uploads/{id}/{file}", logReq(downloadFile))				// Download a shared file
	


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




func logReq(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Generate a token that will permit to prevent unwanted record to database due to browse the upload URL without using the form
	// The trick is that this token is used from an hidden input on the HTML form, and if it's empty it means we're not using the form
	token := generatePassword()
	
	renderTemplate(w, "view.create.file.html", struct {
		TokenAvoidRefresh string
		// Expiration string
	}{
		TokenAvoidRefresh: token,
		// Expiration: time.Now().String(),
	})
}




func viewCreateSecret(w http.ResponseWriter, r *http.Request) {

	// Generate a token that will permit to prevent unwanted record to database due to browse the upload URL without using the form
	// The trick is that this token is used from an hidden input on the HTML form, and if it's empty it means we're not using the form
	token := generatePassword()

	renderTemplate(w, "view.create.secret.html", struct {
		TokenAvoidRefresh string
		// Expiration string
	}{
		TokenAvoidRefresh: token,
		// Expiration: time.Now().String(),
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


		shareContentMap := getShareContent(idToUnlock)
		shareContentType := shareContentMap["type"]
		shareContentValue := shareContentMap["value"]
	

		if givenPasswordHash == sharePasswordHash {
			data := map[string]interface{}{
				"sharePasswordHash": sharePasswordHash,
				"shareContentType": shareContentType,
				"shareContentValue": shareContentValue,
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




func uploadSecret(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	// Ensure that a refresh of the page will not submit a new value in the database
	tokenAvoidRefresh := r.PostFormValue("TokenAvoidRefresh")
	if tokenAvoidRefresh != "" {

		id := uuid.NewString()
		shared_id := uuid.NewString()
		uri := r.Header.Get("Referer")											// Entire path 'http://domain:port/node1/node2/etc.../'
		url := path.Dir(uri)													// Only the 'http://domain:port' part
		link := strings.Join([]string{"/share/", shared_id}, "")


		// Create database entries
		createSecret(id, shared_id, r.PostFormValue("mySecret"), r.PostFormValue("expiration"))


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




func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	// Ensure that a refresh of the page will not submit a new value in the database
	tokenAvoidRefresh := r.PostFormValue("TokenAvoidRefresh")
	if tokenAvoidRefresh != "" {


		id := uuid.NewString()
		shared_id := uuid.NewString()
		uri := r.Header.Get("Referer")											// Entire path 'http://domain:port/node1/node2/etc.../'
		url := path.Dir(uri)													// Only the 'http://domain:port' part
		link := strings.Join([]string{"/share/", shared_id}, "")



		// Get handler for filename, size and headers
		file, handler, err := r.FormFile("myFile")
		if err != nil {
			fmt.Println("Error retrieving the file")
			fmt.Println(err)
			return
		}
		defer file.Close()
		// fmt.Printf("Uploaded file: %+v\n", handler.Filename)
		// fmt.Printf("File size: %+v\n", handler.Size)
		// fmt.Printf("MIME header: %+v\n", handler.Header)

		// Create destination directory root
		dirUploads := "uploads/"
		if _, err := os.Stat(dirUploads); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dirUploads, 0700)
			if err != nil {
				log.Println(err)
			}
		}

		// Create destination directory of the share
		dir := dirUploads + shared_id
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(dir, 0700)
			if err != nil {
				log.Println(err)
			}
		}

		// Create file
		filePath := filepath.Join(dir, filepath.Base(handler.Filename))
		dst, err := os.Create(filePath)
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
		createFile(id, shared_id, filePath, r.PostFormValue("expiration"))		


		
		// Display the confirmation
		renderTemplate(w, "view.confirm.file.html", struct {
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




func downloadFile(w http.ResponseWriter, r *http.Request) {


	url := r.Header.Get("Referer")
	share_id := url[len(url)-36:]	// Just get the last 36 char of the url because the IDs are 36 char length
	shareContentMap := getShareContent(share_id)
	file := shareContentMap["value"]

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Content-Disposition", "attachment; filename=" + file)

    http.ServeFile(w, r, file)
}
