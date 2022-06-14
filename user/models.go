package user

import (
	"fmt"
	"web-chat-go/config"

	_ "github.com/lib/pq"
)

type User struct {
	ID        int
	UserName  string
	SaltPass  string
	Password  []byte
	FirstName string
	LastName  string
}

func AllUsers() (map[string]User, error) {
	db := config.Connect()
	defer db.Close()

	rows, err := db.Query(`
		SELECT
			UserName,
			SaltPass,
			Password,
			FirstName,
			LastName
		FROM users;
	`)
	if err != nil {
		panic(err)
	}

	var dbUsers = make(map[string]User)

	for rows.Next() {
		u := User{}

		err := rows.Scan(
			&u.UserName,
			&u.SaltPass,
			&u.Password,
			&u.FirstName,
			&u.LastName,
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

func SearchUser(un string) (User, error) {
	var u User

	db := config.Connect()
	defer db.Close()

	q := fmt.Sprintf(`
		SELECT
			ID,
			UserName,
			SaltPass,
			Password,
			FirstName,
			LastName
		FROM users
		WHERE
			username = $1;
	`)
	row := db.QueryRow(q, un)

	err := row.Scan(
		&u.ID,
		&u.UserName,
		&u.SaltPass,
		&u.Password,
		&u.FirstName,
		&u.LastName,
	)
	return u, err
}

func AddUser(u User) error {
	db := config.Connect()
	defer db.Close()

	q := fmt.Sprintf(`
		INSERT INTO users (
			UserName,
			SaltPass,
			Password,
			FirstName,
			LastName
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		);
	`)
	_, err := db.Exec(
		q,
		u.UserName, u.SaltPass,
		u.Password, u.FirstName,
		u.LastName,
	)

	return err
}
