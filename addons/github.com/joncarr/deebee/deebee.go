package deebee

import (
	"database/sql"
	"fmt"
	"log"
)

// NewDatabaseConnection returns a pointer to sql.DB
// Parameters:
// ** dbName (string): The name of the database
// ** dbUser (string): Database username
// ** dbPass (string): Database password for the provided user
// ** dbHost (string): Database host name
// sslMode (bool): Determines whether sslmode is turned on or off
// ** It is recommended that you load these variables from environment variables
func NewDatabaseConnection(dbName, dbUser, dbPass, dbHost string, sslMode bool) *sql.DB {

	var connStr string

	if !sslMode {
		connStr = fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", dbName, dbUser, dbPass, dbHost)
	} else {
		connStr = fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=require", dbName, dbUser, dbPass, dbHost)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("Error connecting to database:", err)
		log.Fatal("Terminating due to database connection error.")
	}

	return db
}
