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

	"github.com/ProtonMail/gopenpgp/v3/crypto"

)




var dbFile string = "sqlite.db"




func createDatabase() {

	// first start					=> create db												=> DELETE_DB_ON_NEXT_START = false
	// running without reset		=> do nothing												=> DELETE_DB_ON_NEXT_START = false
	// reset						=> delete then create db (and create if if not exists)		=> DELETE_DB_ON_NEXT_START = true


	// Env var given from pseudo CLI
	var DELETE_DB_ON_NEXT_START, err = strconv.ParseBool(os.Getenv("DELETE_DB_ON_NEXT_START"))
	if err != nil {
		log.Fatal(err)
	}


	var query = `
	CREATE TABLE share (id text not null primary key, pgpkeypublic text, pgpkeyprivate text, password text, maxopen int, currentopen int, expiration text, creation text);
	DELETE FROM share;
	CREATE TABLE file (id text not null primary key, path text, share_id text, FOREIGN KEY(share_id) REFERENCES share(id));
	DELETE FROM file;
	CREATE TABLE secret (id text not null primary key, text text, share_id text, FOREIGN KEY(share_id) REFERENCES share(id));
	DELETE FROM secret;
	`
	
	
	// Reset database only if the user has decided to
	if DELETE_DB_ON_NEXT_START == true {

		// Check if file exists
		if fileExists(dbFile) {
			os.Remove(dbFile)
		}
	
		// Open connexion
		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Create tables
		_, err = db.Exec(query)
		if err != nil {
			log.Printf("%q: %s\n", err, query)
			return
		}

		fmt.Println("Database resetted")
	

	} else {

		// Check if file exists to create it if not
		if ! fileExists(dbFile) {
			
			// Open connexion
			db, err := sql.Open("sqlite3", dbFile)
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()
			
			// Create tables
			_, err = db.Exec(query)
			if err != nil {
				log.Printf("%q: %s\n", err, query)
				return
			}

			fmt.Println("Database created")
		} else {
			fmt.Println("Database found")
		}

	}
}





func createShare(id string, expirationGiven string, maxopenGiven string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()



	t := time.Now()
	now := fmt.Sprintf("%d-%02d-%02dT%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())

	creation := sql.Named("creation", now)
	password := sql.Named("password", generatePassword())
	maxopen := sql.Named("maxopen", maxopenGiven)
	currentopen := 0
	expiration := sql.Named("expiration", expirationGiven)


	pgp := crypto.PGP()
	keyGenHandle := pgp.KeyGeneration().AddUserId("share", id).New()
	keyPrivate, _ := keyGenHandle.GenerateKey()
	keyPublic, _ := keyPrivate.ToPublic()
	keyPrivateChain, _ := keyPrivate.Armor()
	keyPublicChain, _ := keyPublic.GetArmoredPublicKey()


	_, err = db.Exec("INSERT INTO share(id, password, pgpkeypublic, pgpkeyprivate, maxopen, currentopen, expiration, creation) values(:id, :password, :pgpkeypublic, :pgpkeyprivate, :maxopen, :currentopen, :expiration, :creation)", id, password, keyPublicChain, keyPrivateChain, maxopen, currentopen, expiration, creation)
	if err != nil {
		log.Fatal(err)
	}
}




func createFile(id string, share_id string, path string, expiration string, maxopen string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	_, err = db.Exec("INSERT INTO file(id, path, share_id) values(:id, :path, :share_id)", id, path, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(share_id, expiration, maxopen)
}




func createSecret(id string, share_id string, text string, expiration string, maxopen string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO secret(id, text, share_id) values(:id, :text, :share_id)", id, text, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(share_id, expiration, maxopen)
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




// Get the GPG public key of a share
func getShareKeyPublic(share_id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	row := db.QueryRow("SELECT pgpkeypublic FROM share WHERE id = :share_id", share_id)
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




// Get the GPG private key of a share
func getShareKeyPrivate(share_id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	row := db.QueryRow("SELECT pgpkeyprivate FROM share WHERE id = :share_id", share_id)
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




// Get the password of a share
func getShareOpen(share_id string) map[string]string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	row := db.QueryRow("SELECT currentopen, maxopen FROM share WHERE id = :share_id", share_id)
	var rowDataCurrentOpen string
	var rowDataMaxOpen string
	switch err := row.Scan(&rowDataCurrentOpen, &rowDataMaxOpen); err {
		case sql.ErrNoRows:
			fmt.Println("No row returned from table 'share'")
		case nil:
			fmt.Println("Rows found:", rowDataCurrentOpen, "and", rowDataMaxOpen)
		default:
			panic(err)
	}

	
	return map[string]string{
		"currentopen": rowDataCurrentOpen,
		"maxopen": rowDataMaxOpen,
	}

}




// Get the password of a share
func updateShareOpen(share_id string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	row := db.QueryRow("SELECT currentopen FROM share WHERE id = :share_id", share_id)
	var rowDataCurrentOpen string
	switch err := row.Scan(&rowDataCurrentOpen); err {
		case sql.ErrNoRows:
			fmt.Println("No row returned from table 'share'")
		case nil:
			fmt.Println("Rows found:", rowDataCurrentOpen)
		default:
			panic(err)
	}


	// Increment the open (meaning it has been opened one time)
	// currentopen := rowDataCurrentOpen + "1"
	currentopenInt, _ := strconv.Atoi(rowDataCurrentOpen)
	currentopen := currentopenInt + 1

	fmt.Println("rowDataCurrentOpen ", rowDataCurrentOpen)
	fmt.Println("currentopen        ", currentopen)

	_, err = db.Exec("UPDATE share SET currentopen = :currentopen WHERE id = :share_id", currentopen, share_id)
	if err != nil {
		log.Fatal(err)
	}


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




// Get list of shares
func listShareOpen() {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	rows, err := db.Query("SELECT id, creation, expiration FROM share")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var id string
	var creation string
	var expiration string
	for rows.Next() {
		err := rows.Scan(&id, &creation, &expiration)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("ID:" + id + "; Creted:" + creation + "; Expire:" + expiration)

	}
}




// Set a task to run at a specific date
// Regularly check for all shares expiration date, and delete them if expired
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