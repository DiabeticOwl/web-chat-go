package user

import (
	"database/sql"
	"fmt"
)

type User struct {
	ID        int
	UserName  string
	SaltPass  string
	Password  []byte
	FirstName string
	LastName  string
}

func AllUsers(dbConn *sql.DB) (map[string]User, error) {
	rows, err := dbConn.Query(`
		SELECT
			UserName, SaltPass, Password, FirstName, LastName
		FROM users;
	`)
	if err != nil {
		panic(err)
	}

	var dbUsers = make(map[string]User)

	for rows.Next() {
		u := User{}

		err := rows.Scan(
			&u.UserName, &u.SaltPass, &u.Password, &u.FirstName, &u.LastName,
		)
		if err != nil {
			panic(err)
		}

		dbUsers[u.UserName] = u
	}
	if err = rows.Err(); err != nil {
		panic(err)
	}

	return dbUsers, err
}

func IsSigned(u User, err error) bool {
	// Returns whether the User is already in the database.
	return u.UserName != "" && err == nil
}

func SearchUser(dbConn *sql.DB, un string) (User, error) {
	var u User

	q := fmt.Sprintf(`
		SELECT
			ID, UserName, SaltPass, Password, FirstName, LastName
		FROM users
		WHERE
			username = $1;
	`)
	row := dbConn.QueryRow(q, un)

	err := row.Scan(
		&u.ID, &u.UserName, &u.SaltPass, &u.Password, &u.FirstName, &u.LastName,
	)
	return u, err
}

func AddUser(dbConn *sql.DB, u User) error {
	q := fmt.Sprintf(`
		INSERT INTO users (
			UserName, SaltPass, Password, FirstName, LastName
		)
		VALUES (
			$1, $2, $3, $4, $5
		);
	`)
	_, err := dbConn.Exec(
		q, u.UserName, u.SaltPass, u.Password, u.FirstName, u.LastName,
	)

	return err
}
