package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"fmt"
)

func createDatabase() {

	dbFile := "sqlite.db"
	DELETE_DB_ON_NEXT_START := true

	
	if _, err := os.Stat(dbFile); err == nil {
		fmt.Printf("%s found\n", dbFile);

		// Delete database only if the user has decided to.
		if DELETE_DB_ON_NEXT_START == true {
			os.Remove(dbFile)
		}
		
	} else {
		fmt.Printf("creating %s\n", dbFile);
	}



	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	CREATE TABLE share (id integer not null primary key, password text, expiration_datetime datetime);
	DELETE FROM share;
	CREATE TABLE file (id integer not null primary key, path text);
	DELETE FROM file;
	CREATE TABLE secret (id integer not null primary key, text text);
	DELETE FROM secret;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	// tx, err := db.Begin()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer stmt.Close()
	// for i := 0; i < 100; i++ {
	// 	_, err = stmt.Exec(i, fmt.Sprintf("こんにちは世界%03d", i))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
	// err = tx.Commit()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// rows, err := db.Query("select id, name from foo")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()
	// for rows.Next() {
	// 	var id int
	// 	var name string
	// 	err = rows.Scan(&id, &name)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(id, name)
	// }
	// err = rows.Err()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// stmt, err = db.Prepare("select name from foo where id = ?")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer stmt.Close()
	// var name string
	// err = stmt.QueryRow("3").Scan(&name)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(name)

	// _, err = db.Exec("delete from foo")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// _, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// rows, err = db.Query("select id, name from foo")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer rows.Close()
	// for rows.Next() {
	// 	var id int
	// 	var name string
	// 	err = rows.Scan(&id, &name)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Println(id, name)
	// }
	// err = rows.Err()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}




// func createTables() {

// 	db, err := sql.Open("sqlite3", "./foo.db")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	_, err = db.Exec("CREATE TABLE share (id integer not null primary key, password text, expiration_datetime datetime)")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	_, err = db.Exec("CREATE TABLE file (id integer not null primary key, path text)")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	_, err = db.Exec("CREATE TABLE secret (id integer not null primary key, text text)")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }