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

	"crypto/rand"
    "encoding/base64"
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




// func readShare(id sql.NamedArg) {
// 	db, err := sql.Open("sqlite3", dbFile)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()

// 	// openDatabase()


// 	// https://www.calhoun.io/querying-for-a-single-record-using-gos-database-sql-package/
// 	row := db.QueryRow("SELECT password FROM share WHERE :id", id)
// 	var password string
// 	switch err := row.Scan(&password); err {
// 		case sql.ErrNoRows:
// 			fmt.Println("No rows were returned!")
// 		case nil:
// 			fmt.Println(password)
// 		default:
// 			panic(err)
// 	}

// }




// func generatePassword() string {
	
// 	charset := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#^&*()_+-=[]{}|;:,.<>/~"
// 	n := rand.Intn(99-64) + 64

// 	// define 2 vars
// 	// the first is just a slice of rune that is n length
// 	// the second is a slice of rune seeded from random string source

// 	s, r := make([]rune, n), []rune(charset)

// 	// loop through the empty rune slice
// 	for i := range s {
// 		// generate a prime number based on the given bit length of r log2(n)+1
// 		// for example, 12 would return a bit-length of 4 so the prime number would be based on 4
// 		p, _ := rand.Prime(rand.Reader, len(r))

// 		// define 2 additional variables,
// 		// x is based on the Unit64 representation of p from above
// 		// y is based on the uint64 type case of the length of R (our []rune(randomString))
// 		x, y := p.Uint64(), uint64(len(r)) // note: uint64 here because we know it will not be negative

// 		// finally for the index of if in s which is just an empty slice of rune
// 		// choose a  rune from r where the index is the result of modulus operationx x%y

// 		s[i] = r[x%y]
// 	}

// 	// after we finish looping through the rune and assigning values to each index,
// 	// return the string
// 	return string(s)
// }




func generatePassword() string {


	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
	   panic(err)
	}

	password := base64.StdEncoding.EncodeToString(b)
	
	fmt.Println(password)

	return password
 }




func createShare(id string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// openDatabase()



	// rint := rand.Intn(99-64) + 64
	// rpwd, err := generatePassword()
	// if err != nil {
	//   log.Fatal(err)
	// }
	// log.Printf(rpwd)

	




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
			fmt.Println(password)
		default:
			panic(err)
	}
	
	return password
}