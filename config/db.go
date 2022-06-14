package config

import (
	"database/sql"
	"os"
)

func Connect() *sql.DB {
	db, err := sql.Open(
		"postgres",
		os.Getenv("POSTGRECRED"),
	)
	if err != nil {
		panic(err)
	}

	return db
}
