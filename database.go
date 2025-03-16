package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strconv"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/go-co-op/gocron"

)




var dbFile string = "sqlite.db"




func createDatabase() {

	// Env var given from pseudo CLI
	var DELETE_DB_ON_NEXT_START, err = strconv.ParseBool(os.Getenv("DELETE_DB_ON_NEXT_START"))
	if err != nil {
		log.Fatal(err)
	}


	if _, err := os.Stat(dbFile); err == nil {
		fmt.Printf("%s found\n", dbFile);

		// Delete database only if the user has decided to.
		if DELETE_DB_ON_NEXT_START == true {
			os.Remove(dbFile)
			db, err := sql.Open("sqlite3", dbFile)
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

		
			sqlStmt := `
			CREATE TABLE share (id text not null primary key, password text, maxopen int, expiration text, creation text);
			DELETE FROM share;
			CREATE TABLE file (id text not null primary key, path text, share_id text, FOREIGN KEY(share_id) REFERENCES share(id));
			DELETE FROM file;
			CREATE TABLE secret (id text not null primary key, text text, share_id text, FOREIGN KEY(share_id) REFERENCES share(id));
			DELETE FROM secret;
			`
			_, err = db.Exec(sqlStmt)
			if err != nil {
				log.Printf("%q: %s\n", err, sqlStmt)
				return
			}
		}
		
	}


}




func createShare(id string, expirationChosen string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()



	t := time.Now()
	now := fmt.Sprintf("%d-%02d-%02dT%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())

	creation := sql.Named("creation", now)
	password := sql.Named("password", generatePassword())
	maxopen := 3
	expiration := sql.Named("expiration", expirationChosen)


	_, err = db.Exec("INSERT INTO share(id, password, maxopen, expiration, creation) values(:id, :password, :maxopen, :expiration, :creation)", id, password, maxopen, expiration, creation)
	if err != nil {
		log.Fatal(err)
	}
}




func createFile(id string, share_id string, path string, expiration string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	_, err = db.Exec("INSERT INTO file(id, path, share_id) values(:id, :path, :share_id)", id, path, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(share_id, expiration)
}




func createSecret(id string, share_id string, text string, expiration string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO secret(id, text, share_id) values(:id, :text, :share_id)", id, text, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(share_id, expiration)
}




// Get the content of a share
func getShareContent(share_id string) map[string]string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()



	rowSecret := db.QueryRow("SELECT text FROM secret WHERE share_id = :share_id", share_id)
	var secretText string
	switch err := rowSecret.Scan(&secretText); err {
		case sql.ErrNoRows:
			fmt.Println("No row returned from table 'secret'")
		case nil:
			fmt.Println("Row found:", secretText)
		default:
			panic(err)
	}



	rowFile := db.QueryRow("SELECT path FROM file WHERE share_id = :share_id", share_id)
	var filePath string
	switch err := rowFile.Scan(&filePath); err {
		case sql.ErrNoRows:
			fmt.Println("No row returned from table 'file'")
		case nil:
			fmt.Println("Row found:", filePath)
		default:
			panic(err)
	}
	

	if secretText != "" {
		return map[string]string{
			"type": "secret",
			"value": secretText,
		}

	} else if filePath != ""  {
		return map[string]string{
			"type": "file",
			"value": filePath,
		}

	} else {
		return map[string]string{
			"type": "none",
			"value": "none",
		}
	}
}




// Get the password of a share
func getSharePassword(share_id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	row := db.QueryRow("SELECT password FROM share WHERE id = :share_id", share_id)
	var rowData string
	switch err := row.Scan(&rowData); err {
		case sql.ErrNoRows:
			fmt.Println("No row returned from table 'share'")
		case nil:
			fmt.Println("Row found:", rowData)
		default:
			panic(err)
	}
	
	return rowData
}




// Delete a share and also its related secrets and files (and delete file from filesystem aswell)
func deleteShare(share_id string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	rowShare := db.QueryRow("DELETE FROM share WHERE id = :share_id", share_id)
	var rowShareData string
	switch err := rowShare.Scan(&rowShareData); err {
		case sql.ErrNoRows:
			fmt.Println("Row deleted from table 'share'")
		// case nil:
		// 	fmt.Println("Row found:", rowShareData)
		default:
			panic(err)
	}


	rowSecret := db.QueryRow("DELETE FROM secret WHERE share_id = :share_id", share_id)
	var rowSecretData string
	switch err := rowSecret.Scan(&rowSecretData); err {
		case sql.ErrNoRows:
			fmt.Println("Row deleted from table 'secret'")
		// case nil:
		// 	fmt.Println("Row found:", rowSecretData)
		default:
			panic(err)
	}


	rowFile := db.QueryRow("DELETE FROM file WHERE share_id = :share_id", share_id)
	var rowFileData string
	switch err := rowFile.Scan(&rowFileData); err {
		case sql.ErrNoRows:
			fmt.Println("Row deleted from table 'file'")
		// case nil:
		// 	fmt.Println("Row found:", rowFileData)
		default:
			panic(err)
	}


	// Delete the directory containing files of the share
	deletePath("uploads/" + share_id)

}




// Set a task to run at a specific date
// Regularly check for all share expiration date, and delete them if expired
func periodicClean() {

	task := gocron.NewScheduler(time.UTC)
	task.Every(1).Minutes().Do(func() {
		fmt.Println("Periodic cleaning task started at:", time.Now())

		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	
		rows, err := db.Query("SELECT id, expiration FROM share")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var rowDataId string
			var rowDataExpiration string

			err := rows.Scan(&rowDataId, &rowDataExpiration)
			if err != nil {
				log.Fatal(err)
			}


			now := time.Now()
			timeLayout := "2006-01-02T15:04"
			expiration, err := time.Parse(timeLayout, rowDataExpiration)
			if err != nil {
				log.Fatal(err)
			}


			// Delete share if its expiration date is before now
			if now.After(expiration) {
				go deleteShare(rowDataId)	// Set as Goroutine to avoid database crash due to too many connexion opened
			}

			// if now.After(expiration) {
			// 	fmt.Println()
			// 	fmt.Println("EXPIRED")
			// 	fmt.Println("id            ", rowDataId)
			// 	fmt.Println("expiration row", rowDataExpiration)
			// 	fmt.Println("expiration    ", expiration)
			// 	fmt.Println("now           ", now)

			// } else if now.Before(expiration)  {
				
			// 	fmt.Println()
			// 	fmt.Println("ALIVE")
			// 	fmt.Println("id            ", rowDataId)
			// 	fmt.Println("expiration row", rowDataExpiration)
			// 	fmt.Println("expiration    ", expiration)
			// 	fmt.Println("now           ", now)

			// }
		}
		
		// fmt.Println("-------------")

    })

    task.StartAsync()

    // Prevent exit
    select {}
}