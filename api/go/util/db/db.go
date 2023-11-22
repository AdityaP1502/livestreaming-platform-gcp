package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AdityaP1502/livestreaming-platform-gcp/api/go/base"
	_ "github.com/go-sql-driver/mysql"
)

func OpenDatabase(dbUsername string, dbPassword string, dbInstanceAddress string,
	dbInstancePort int, dbName string) base.App {

	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s", dbUsername, dbPassword,
		dbInstanceAddress, dbInstancePort, dbName,
	))

	if err != nil {
		log.Fatal(err)
	}

	return base.App{Connection: db}
}

func CloseDatabase(db base.App) error {
	return db.Connection.Close()
}

func CheckIfRowExist(err error) (bool, error) {
	if err == nil {
		return true, nil
	}

	if err == sql.ErrNoRows {
		return false, nil
	}

	return false, err
}
