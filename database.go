package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"fmt"
	"time"

	// "github.com/google/uuid"
	// "github.com/sethvargo/go-password/password"

	// "crypto/rand"
    // "encoding/base64"
	"strconv"
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
		
			// openDatabase()

		
			sqlStmt := `
			CREATE TABLE share (id text not null primary key, password text, maxopen int, expiration datetime, creation datetime);
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



func createShare(id string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()



	// id := sql.Named("id", uuid.NewString())
	password := sql.Named("password", generatePassword())
	maxopen := 3
	expiration := sql.Named("datetime", time.Now())
	creation := sql.Named("datetime", time.Now())


	_, err = db.Exec("INSERT INTO share(id, password, maxopen, expiration, creation) values(:id, :password, :maxopen, :datetime, :datetime)", id, password, maxopen, expiration, creation)
	if err != nil {
		log.Fatal(err)
	}


	// Return the ID to be able to read it just after the creation
	// return id
}




func createFile(id string, share_id string, path string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// openDatabase()


	// id := sql.Named("id", uuid.NewString())
	// share_id := "ok"


	_, err = db.Exec("INSERT INTO file(id, path, share_id) values(:id, :path, :share_id)", id, path, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(share_id)
	// readShare(createShare())
}




func createSecret(id string, share_id string, text string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// openDatabase()


	// id := sql.Named("id", uuid.NewString())
	// share_id := "ok"


	_, err = db.Exec("INSERT INTO secret(id, text, share_id) values(:id, :text, :share_id)", id, text, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(share_id)
	// readShare(createShare())
}




// func readSecret(id string) string {
// 	db, err := sql.Open("sqlite3", dbFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
// 	row := db.QueryRow("SELECT text FROM secret WHERE id = :id", id)
// 	var text string
// 	switch err := row.Scan(&text); err {
// 		case sql.ErrNoRows:
// 			fmt.Println("No rows were returned!")
// 		case nil:
// 			fmt.Println(text)
// 		default:
// 			panic(err)
// 	}
	
// 	return text
// }




// func readFile(id string) string {
// 	db, err := sql.Open("sqlite3", dbFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()


// 	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
// 	row := db.QueryRow("SELECT path FROM file WHERE id = :id", id)
// 	var path string
// 	switch err := row.Scan(&path); err {
// 		case sql.ErrNoRows:
// 			fmt.Println("No rows were returned!")
// 		case nil:
// 			fmt.Println(path)
// 		default:
// 			panic(err)
// 	}
	
// 	return path
// }




// Get the content of a share
func getShareContent(share_id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()



	row1 := db.QueryRow("SELECT text FROM secret where share_id = :share_id", share_id)
	var rowSecret string
	switch err := row1.Scan(&rowSecret); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println("Row found:", rowSecret)
		default:
			panic(err)
	}



	row2 := db.QueryRow("SELECT path FROM file where share_id = :share_id", share_id)
	var rowFile string
	switch err := row2.Scan(&rowFile); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println("Row found:", rowFile)
		default:
			panic(err)
	}
	

	var shareContent string
	if rowSecret != "" {
		shareContent = rowSecret

		fmt.Printf("secret", shareContent)

	} else if rowFile != ""  {
		shareContent = rowFile
		fmt.Printf("file", shareContent)

	} else {
		shareContent = "empty"
		fmt.Printf("empty", shareContent)
	}


	return shareContent
}




// Get the password of a share
func getSharePassword(share_id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
	row := db.QueryRow("SELECT password FROM share WHERE id = :share_id", share_id)
	var rowData string
	switch err := row.Scan(&rowData); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println("Row found:", rowData)
		default:
			panic(err)
	}
	
	return rowData
}


