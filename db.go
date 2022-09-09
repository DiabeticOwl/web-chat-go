package main

import (
	"database/sql"
	"os"

	// Register of the two SQL drivers.
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func connect() *sql.DB {
	driverName := "sqlite3"
	dataSourceName := "./db.db"

	if !debug {
		driverName = "postgres"
		dataSourceName = os.Getenv("POSTGRECRED")
	}

	// Open will register the desired driver and open a connection with
	// the given dataSourceName.
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}

	return db
}
