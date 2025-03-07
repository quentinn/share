package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"fmt"
	"time"

	"github.com/satori/go.uuid"
)




var dbFile string = "sqlite.db"
var DELETE_DB_ON_NEXT_START bool = false




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
		
			sqlStmt := `
			CREATE TABLE share (id text not null primary key, password text, expiration datetime);
			DELETE FROM share;
			CREATE TABLE file (id text not null primary key, path text);
			DELETE FROM file;
			CREATE TABLE secret (id text not null primary key, text text);
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




func createShare() {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	id := sql.Named("id", uuid.NewV4())
	password := sql.Named("password", uuid.NewV4())
	datetime := sql.Named("datetime", time.Now())


	_, err = db.Exec("INSERT INTO share(id, password, expiration) values(:id, :password, :datetime)", id, password, datetime)
	if err != nil {
		log.Fatal(err)
	}
}




func createFile(path string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	id := sql.Named("id", uuid.NewV4())


	_, err = db.Exec("INSERT INTO file(id, path) values(:id, :path)", id, path)
	if err != nil {
		log.Fatal(err)
	}
}