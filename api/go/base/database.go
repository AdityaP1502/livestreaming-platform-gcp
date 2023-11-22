package base

import "database/sql"

type App struct {
	Connection *sql.DB
}
