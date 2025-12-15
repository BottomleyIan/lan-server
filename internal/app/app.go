package app

import (
	"database/sql"

	"bottomley.ian/musicserver/internal/db"
	"bottomley.ian/musicserver/internal/services/fs"
)

type App struct {
	DB      *sql.DB
	Queries *db.Queries
	FS      fs.FS
}
