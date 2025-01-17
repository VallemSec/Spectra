package parser

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func Connect(username, password, host, port, dbname string) (*sql.DB, error) {
	// connect to the mysql database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbname))
	if err != nil {
		return nil, err
	}

	// check if the connection is working
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GenerateConn(username, password, host, port, dbname string) string {
	return fmt.Sprintf("%s:%s@%s:%s/%s", username, password, host, port, dbname)
}

func InsertResult(db *sql.DB, result string) (string, error) {
	uuidNew := uuid.New()

	uuidString := "parser-" + uuidNew.String()

	_, err := db.Exec("INSERT INTO key_value (`key`, value) VALUES (?, ?)", uuidString, result)
	if err != nil {
		return "", err
	}

	return uuidString, nil
}
