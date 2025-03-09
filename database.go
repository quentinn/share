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
)




var dbFile string = "sqlite.db"
var DELETE_DB_ON_NEXT_START bool = false




// func openDatabase() {
// 	db, err := sql.Open("sqlite3", dbFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	return db
// }




func createDatabase() {

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




func createFile(id string, path string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// openDatabase()


	// id := sql.Named("id", uuid.NewString())
	share_id := "ok"


	_, err = db.Exec("INSERT INTO file(id, path, share_id) values(:id, :path, :share_id)", id, path, share_id)
	if err != nil {
		log.Fatal(err)
	}


	createShare(id)
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




func readSecret(id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
	row := db.QueryRow("SELECT text FROM secret WHERE id = :id", id)
	var text string
	switch err := row.Scan(&text); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println(text)
		default:
			panic(err)
	}
	
	return text
}




func readFile(id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
	row := db.QueryRow("SELECT path FROM file WHERE id = :id", id)
	var path string
	switch err := row.Scan(&path); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println(path)
		default:
			panic(err)
	}
	
	return path
}




func readShare(id string) string {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
	row := db.QueryRow("SELECT password FROM share WHERE id = :id", id)
	var password string
	switch err := row.Scan(&password); err {
		case sql.ErrNoRows:
			fmt.Println("No rows were returned!")
		case nil:
			fmt.Println("Row found:", password)
		default:
			panic(err)
	}
	
	return password
}


// SELECT id, text FROM secret
// where share_id = "8ff11545-3b0f-4c87-82d6-d0635238fa83"
// UNION
// SELECT id, path FROM file
// where share_id = "8ff11545-3b0f-4c87-82d6-d0635238fa83"