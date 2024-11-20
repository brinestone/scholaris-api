package institutions

import (
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("institution_db", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})
