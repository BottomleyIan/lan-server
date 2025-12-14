package app

import (
	"database/sql"

	"bottomley.ian/musicserver/internal/db"
)

type App struct {
	DB      *sql.DB
	Queries *db.Queries
}
